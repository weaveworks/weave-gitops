package profiles

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

type UpdateOptions struct {
	Namespace       string
	Port            string
	ProfileFilepath string
	ProfileName     string
	ProfileVersion  string
	ClientSet       kubernetes.Interface
	Writer          io.Writer
}

func UpdateProfile(ctx context.Context, opts UpdateOptions) error {
	fmt.Fprintf(opts.Writer, "Updating profile %s to version %s", opts.ProfileName, opts.ProfileVersion)

	helmRelease := &helmv2.HelmRelease{}
	content, err := ioutil.ReadFile(opts.ProfileFilepath)

	if err != nil {
		return fmt.Errorf(`error reading profile %q at path %q: %w`, opts.ProfileName, opts.ProfileFilepath, err)
	}

	err = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(content), 4096).Decode(helmRelease)
	if err != nil {
		return fmt.Errorf("error unmarshaling %q: %w", opts.ProfileFilepath, err)
	}

	profiles, err := getProfiles(ctx, GetOptions{
		Namespace: opts.Namespace,
		Port:      opts.Port,
		ClientSet: opts.ClientSet,
	})

	if err != nil {
		return fmt.Errorf("failed to get available profiles: %w", err)
	}

	profileExists := false

	for _, profile := range profiles.Profiles {
		if profile.Name == opts.ProfileName {
			profileExists = true

			if !containsString(profile.AvailableVersions, opts.ProfileVersion) {
				return fmt.Errorf("version %q is not available for profile %q. Available versions: %s", opts.ProfileVersion, opts.ProfileName, strings.Join(profile.AvailableVersions, ","))
			}
		}
	}

	if !profileExists {
		return fmt.Errorf("profile %q does not exist. Run \"gitops get profiles\" to see available profiles", opts.ProfileName)
	}

	helmRelease.Spec.Chart.Spec.Version = opts.ProfileVersion

	return writeResource(helmRelease, opts.ProfileFilepath)
}

func containsString(list []string, elem string) bool {
	for _, i := range list {
		if i == elem {
			return true
		}
	}

	return false
}

func writeResource(obj runtime.Object, filename string) error {
	e := kjson.NewSerializerWithOptions(kjson.DefaultMetaFactory, nil, nil, kjson.SerializerOptions{Yaml: true, Strict: true})
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	if err := e.Encode(obj, f); err != nil {
		return err
	}

	return nil
}
