package install

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const FluentBitHRName = "fluent-bit"

const configInputs = `
[INPUT]
    Name tail
    Path /var/log/containers/*.log
    multiline.parser docker, cri
    Tag kube.*
    Mem_Buf_Limit 5MB
    Skip_Long_Lines Off
`

const configFilters = `
[FILTER]
    Name kubernetes
    Match kube.*
    Merge_Log On
    Keep_Log Off
    K8S-Logging.Parser On
    K8S-Logging.Exclude On
[FILTER]
    Name    grep
    Match   *
    Exclude $kubernetes['namespace_name'] (gitops-run|kube-system)
[FILTER]
    Name    grep
    Match   *
    Exclude $kubernetes['pod_name'] ^fluent\-bit
`

func makeConfigOutputs(bucket string, port int32) (string, error) {
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("port %d not between 1 and 65535", port)
	}

	const configOutputs = `
[OUTPUT]
    Name s3
    Match kube.*
    bucket {{ .Bucket }}
    endpoint http://run-dev-bucket.gitops-run.svc:{{ .Port }}
    tls Off
    tls.verify Off
    use_put_object true
    preserve_data_ordering true
    static_file_path true
    total_file_size 1M
    upload_timeout 15s
    s3_key_format /fluent-bit-logs/$TAG[4].%Y%m%d%H%M%S
`
	tmpl, err := template.New("configOutputs").Parse(strings.TrimSpace(configOutputs))
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}
	var result strings.Builder
	err = tmpl.Execute(&result, struct {
		Bucket string
		Port   int32
	}{Bucket: bucket, Port: port})
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	return result.String(), nil
}

func mapToJSON(m map[string]interface{}) (*v1.JSON, error) {
	// convert the map to JSON
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	// create a new JSON struct
	result := &v1.JSON{Raw: b}
	return result, nil
}

func makeFluentBitHelmRepository(namespace string) *sourcev1.HelmRepository {
	helmRepository := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluent",
			Namespace: namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: "https://fluent.github.io/helm-charts",
		},
	}

	return helmRepository
}

func makeFluentBitHelmRelease(name, fluxNamespace, targetNamespace, bucketName string, bucketServerPort int32) (*helmv2.HelmRelease, error) {
	configOutputs, err := makeConfigOutputs(bucketName, bucketServerPort)
	if err != nil {
		return nil, err
	}

	values, err := mapToJSON(map[string]interface{}{
		"env": []map[string]interface{}{
			{
				"name": "AWS_ACCESS_KEY_ID",
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": constants.RunDevBucketCredentials,
						"key":  "accesskey",
					},
				},
			},
			{
				"name": "AWS_SECRET_ACCESS_KEY",
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": constants.RunDevBucketCredentials,
						"key":  "secretkey",
					},
				},
			},
		},
		"config": map[string]interface{}{
			"inputs":  strings.TrimSpace(configInputs),
			"filters": strings.TrimSpace(configFilters),
			"outputs": configOutputs,
		},
	})
	if err != nil {
		return nil, err
	}

	obj := helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: fluxNamespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "fluent-bit",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      "HelmRepository",
						Name:      "fluent",
						Namespace: fluxNamespace,
					},
					Version: "*",
				},
			},
			Interval:        metav1.Duration{Duration: 1 * time.Hour},
			ReleaseName:     name,
			TargetNamespace: targetNamespace,
			Install: &helmv2.Install{
				CRDs: helmv2.Create,
			},
			Upgrade: &helmv2.Upgrade{
				CRDs: helmv2.CreateReplace,
			},
			Values: values,
		},
	}

	return &obj, nil
}

func UninstallFluentBit(ctx context.Context, log logger.Logger, kubeClient client.Client, hrNamespace, hrName string) error {
	hr := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hrName,
			Namespace: hrNamespace,
		},
	}

	log.Actionf("Removing Fluent Bit HelmRelease %s/%s ...", hr.Namespace, hr.Name)

	if err := kubeClient.Delete(ctx, hr); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete HelmRelease: %w", err)
		}
	}

	log.Actionf("Waiting for HelmRelease %s/%s to be deleted...", hr.Namespace, hr.Name)

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 4 * time.Second,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
	}, func() (done bool, err error) {
		if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(hr), hr); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			} else {
				return false, fmt.Errorf("failed retrieving HelmRelease: %w", err)
			}
		}
		return false, nil
	}); err != nil {
		return fmt.Errorf("failed waiting for HelmRelease %s/%s to be deleted: %w", hr.Namespace, hr.Name, err)
	}

	log.Successf("HelmRelease %s/%s deleted", hr.Namespace, hr.Name)

	helmRepository := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluent",
			Namespace: hrNamespace,
		},
	}

	if err := kubeClient.Delete(ctx, helmRepository); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete HelmRepository: %w", err)
		}
	}

	log.Successf("HelmRepository %s/%s deleted", helmRepository.Namespace, helmRepository.Name)

	return nil
}

func InstallFluentBit(ctx context.Context, log logger.Logger, kubeClient client.Client, fluxNamespace, targetNamespace, name, bucketName string, bucketServerPort int32) error {
	helmRepo := makeFluentBitHelmRepository(fluxNamespace)

	log.Actionf("creating HelmRepository %s/%s", helmRepo.Namespace, helmRepo.Name)

	if err := kubeClient.Create(ctx, helmRepo); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// do nothing
		} else {
			return err
		}
	}

	log.Actionf("creating HelmRelease %s/%s", fluxNamespace, name)

	helmRelease, err := makeFluentBitHelmRelease(name, fluxNamespace, targetNamespace, bucketName, bucketServerPort)
	if err != nil {
		return err
	}

	log.Actionf("creating HelmRelease %s/%s", helmRelease.Namespace, helmRelease.Name)

	if err := kubeClient.Create(ctx, helmRelease); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// do nothing
		} else {
			return err
		}
	}

	log.Actionf("waiting for HelmRelease %s/%s to be ready", helmRelease.Namespace, helmRelease.Name)

	//nolint:staticcheck
	if err := wait.Poll(2*time.Second, 5*time.Minute, func() (bool, error) {
		instance := appsv1.DaemonSet{}
		if err := kubeClient.Get(
			ctx,
			types.NamespacedName{
				Name:      name,
				Namespace: targetNamespace,
			}, &instance); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			} else {
				return false, err
			}
		}

		if instance.Status.NumberReady >= 1 {
			return true, nil
		}

		return false, nil
	}); err != nil {
		log.Failuref("HelmRelease %s/%s failed to become ready", helmRelease.Namespace, helmRelease.Name)
		return err
	}

	log.Successf("HelmRelease %s/%s is ready", helmRelease.Namespace, helmRelease.Name)
	return nil
}
