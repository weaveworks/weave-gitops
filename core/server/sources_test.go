package server

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateHelmRepository(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmRepositoryReq{
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "someapp",
			Url:       "someurl",
			Interval:  &pb.Interval{Minutes: 1},
		}

		res, err := c.AddHelmRepository(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmRepository{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmRepository(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmRepository{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmRepositoryReq{
			Name:      "mykustomization",
			Namespace: ns.Name,
			AppName:   "",
			Url:       "",
			Interval:  nil,
		}

		res, err := c.AddHelmRepository(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmRepository{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmRepository(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmRepository{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListHelmRepositories(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddHelmRepositoryReq{
		Name:      "myhelmrepository",
		Namespace: ns.Name,
		AppName:   appName,
		Url:       "someurl",
		Interval:  &pb.Interval{Minutes: 1},
	}

	addRes, err := c.AddHelmRepository(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedKustomizationReq := &pb.AddHelmRepositoryReq{
		Name:      "otherhelmrepository",
		Namespace: ns.Name,
		AppName:   "",
		Url:       "someurl",
		Interval:  &pb.Interval{Minutes: 1},
	}

	_, err = c.AddHelmRepository(ctx, unAssociatedKustomizationReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoryReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmRepositories).To(HaveLen(1))
	g.Expect(res.HelmRepositories[0].Name).To(Equal(r.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListHelmRepositories(ctx, &pb.ListHelmRepositoryReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.HelmRepositories).To(HaveLen(2))
}

func TestCreateHelmChart(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmChartReq{
			AppName:   "someapp",
			Namespace: ns.Name,
			HelmChart: &pb.HelmChart{
				Name:      "myhelmchart",
				Namespace: ns.Name,
				SourceRef: &pb.SourceRef{
					Kind: pb.SourceRef_HelmRepository,
					Name: "myhelmrepository",
				},
				Chart:    "mychart",
				Version:  "v0.0.0",
				Interval: &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddHelmChart(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmChart{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.HelmChart.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmChart(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmChart{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmChartReq{
			AppName:   "",
			Namespace: ns.Name,
			HelmChart: &pb.HelmChart{
				Name:      "myhelmchart",
				Namespace: ns.Name,
				SourceRef: &pb.SourceRef{
					Kind: pb.SourceRef_HelmRepository,
					Name: "myhelmrepository",
				},
				Chart:    "mychart",
				Version:  "v0.0.0",
				Interval: &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddHelmChart(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.HelmChart{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.HelmChart.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmChart(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.HelmChart{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListHelmCharts(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddHelmChartReq{
		AppName:   appName,
		Namespace: ns.Name,
		HelmChart: &pb.HelmChart{
			Name:      "myhelmchart",
			Namespace: ns.Name,
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_HelmRepository,
				Name: "myhelmrepository",
			},
			Chart:    "mychart",
			Version:  "v0.0.0",
			Interval: &pb.Interval{Minutes: 1},
		},
	}

	addRes, err := c.AddHelmChart(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedHelmChartReq := &pb.AddHelmChartReq{
		AppName:   "",
		Namespace: ns.Name,
		HelmChart: &pb.HelmChart{
			Name:      "otherhelmrepository",
			Namespace: ns.Name,
			SourceRef: &pb.SourceRef{
				Kind: pb.SourceRef_HelmRepository,
				Name: "myhelmrepository",
			},
			Chart:    "mychart",
			Version:  "v0.0.0",
			Interval: &pb.Interval{Minutes: 1},
		},
	}

	_, err = c.AddHelmChart(ctx, unAssociatedHelmChartReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListHelmCharts(ctx, &pb.ListHelmChartReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmCharts).To(HaveLen(1))
	g.Expect(res.HelmCharts[0].Name).To(Equal(r.HelmChart.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListHelmCharts(ctx, &pb.ListHelmChartReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.HelmCharts).To(HaveLen(2))
}

func TestCreateBucket(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddBucketReq{
			AppName:   "someapp",
			Namespace: ns.Name,
			Bucket: &pb.Bucket{
				Name:          "mybucket",
				Namespace:     ns.Name,
				Endpoint:      "endpoint",
				Insecure:      true,
				Provider:      pb.Bucket_AWS,
				Region:        "myregion",
				SecretRefName: "secret_name",
				Interval:      &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddBucket(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.Bucket{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Bucket.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToBucket(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.Bucket{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddBucketReq{
			AppName:   "",
			Namespace: ns.Name,
			Bucket: &pb.Bucket{
				Name:          "mybucket",
				Namespace:     ns.Name,
				Endpoint:      "endpoint",
				Insecure:      true,
				Provider:      pb.Bucket_AWS,
				Region:        "myregion",
				SecretRefName: "secret_name",
				Interval:      &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddBucket(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &sourcev1beta1.Bucket{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.Bucket.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToBucket(r)

		opt := cmpopts.IgnoreFields(sourcev1beta1.Bucket{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListBuckets(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddBucketReq{
		AppName:   appName,
		Namespace: ns.Name,
		Bucket: &pb.Bucket{
			Name:          "mybucket",
			Namespace:     ns.Name,
			Endpoint:      "endpoint",
			Insecure:      true,
			Provider:      pb.Bucket_AWS,
			Region:        "myregion",
			SecretRefName: "secret_name",
			Interval:      &pb.Interval{Minutes: 1},
		},
	}

	addRes, err := c.AddBucket(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedBucketReq := &pb.AddBucketReq{
		AppName:   "",
		Namespace: ns.Name,
		Bucket: &pb.Bucket{
			Name:          "othermybucket",
			Namespace:     ns.Name,
			Endpoint:      "endpoint",
			Insecure:      true,
			Provider:      pb.Bucket_AWS,
			Region:        "myregion",
			SecretRefName: "secret_name",
			Interval:      &pb.Interval{Minutes: 1},
		},
	}

	_, err = c.AddBucket(ctx, unAssociatedBucketReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListBuckets(ctx, &pb.ListBucketReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.Buckets).To(HaveLen(1))
	g.Expect(res.Buckets[0].Name).To(Equal(r.Bucket.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListBuckets(ctx, &pb.ListBucketReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.Buckets).To(HaveLen(2))
}

func TestCreateHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("with app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmReleaseReq{
			AppName:   "someapp",
			Namespace: ns.Name,
			HelmRelease: &pb.HelmRelease{
				Name:        "myhelmrelease",
				Namespace:   ns.Name,
				ReleaseName: "rname",
				HelmChart: &pb.HelmChart{
					Name:      "mychart",
					Namespace: ns.Name,
					Chart:     "chart0",
					Version:   "v0.0.0",
					Interval:  &pb.Interval{Minutes: 1},
					SourceRef: &pb.SourceRef{
						Kind: pb.SourceRef_HelmRepository,
						Name: "myhelmrepository",
					},
				},
				Interval: &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddHelmRelease(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &helmv2beta1.HelmRelease{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.HelmRelease.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmRelease(r)

		opt := cmpopts.IgnoreFields(helmv2beta1.HelmRelease{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}
	})

	t.Run("no app association", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.AddHelmReleaseReq{
			AppName:   "",
			Namespace: ns.Name,
			HelmRelease: &pb.HelmRelease{
				Name:        "myhelmrelease",
				Namespace:   ns.Name,
				ReleaseName: "rname",
				HelmChart: &pb.HelmChart{
					Name:      "mychart",
					Namespace: ns.Name,
					Chart:     "chart0",
					Version:   "v0.0.0",
					Interval:  &pb.Interval{Minutes: 1},
					SourceRef: &pb.SourceRef{
						Kind: pb.SourceRef_HelmRepository,
						Name: "myhelmrepository",
					},
				},
				Interval: &pb.Interval{Minutes: 1},
			},
		}

		res, err := c.AddHelmRelease(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(res.Success).To(BeTrue())

		actual := &helmv2beta1.HelmRelease{}

		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.HelmRelease.Name, Namespace: ns.Name}, actual)).To(Succeed())

		expected := stypes.ProtoToHelmRelease(r)

		opt := cmpopts.IgnoreFields(helmv2beta1.HelmRelease{}, diffIgnoredFields...)
		diff := cmp.Diff(*actual, expected, opt)

		if diff != "" {
			t.Error(fmt.Errorf("(-actual +expected):\n%s", diff))
		}

		g.Expect(actual.Labels["app.kubernetes.io/part-of"]).To(Equal(""))
	})
}

func TestListHelmReleases(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	appName := "myapp"
	ns := newNamespace(ctx, k, g)

	r := &pb.AddHelmReleaseReq{
		AppName:   appName,
		Namespace: ns.Name,
		HelmRelease: &pb.HelmRelease{
			Name:        "myhelmrelease",
			Namespace:   ns.Name,
			ReleaseName: "rname",
			HelmChart: &pb.HelmChart{
				Name:      "mychart",
				Namespace: ns.Name,
				Chart:     "chart0",
				Version:   "v0.0.0",
				Interval:  &pb.Interval{Minutes: 1},
				SourceRef: &pb.SourceRef{
					Kind: pb.SourceRef_HelmRepository,
					Name: "myhelmrepository",
				},
			},
			Interval: &pb.Interval{Minutes: 1},
		},
	}

	addRes, err := c.AddHelmRelease(ctx, r)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	unAssociatedHelmReleaseReq := &pb.AddHelmReleaseReq{
		AppName:   "",
		Namespace: ns.Name,
		HelmRelease: &pb.HelmRelease{
			Name:        "myotherhelmrelease",
			Namespace:   ns.Name,
			ReleaseName: "rname",
			HelmChart: &pb.HelmChart{
				Name:      "mychart",
				Namespace: ns.Name,
				Chart:     "chart0",
				Version:   "v0.0.0",
				Interval:  &pb.Interval{Minutes: 1},
				SourceRef: &pb.SourceRef{
					Kind: pb.SourceRef_HelmRepository,
					Name: "myhelmrepository",
				},
			},
			Interval: &pb.Interval{Minutes: 1},
		},
	}

	_, err = c.AddHelmRelease(ctx, unAssociatedHelmReleaseReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(addRes.Success).To(BeTrue())

	res, err := c.ListHelmReleases(ctx, &pb.ListHelmReleaseReq{
		AppName:   appName,
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res.HelmReleases).To(HaveLen(1))
	g.Expect(res.HelmReleases[0].Name).To(Equal(r.HelmRelease.Name))

	// Ensure our filtering logic is working for `AppName`
	all, err := c.ListHelmReleases(ctx, &pb.ListHelmReleaseReq{
		AppName:   "",
		Namespace: ns.Name,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(all.HelmReleases).To(HaveLen(2))
}

func TestGetAllResourcesByLabel(t *testing.T) {
	//g := NewGomegaWithT(t)
	//
	//ctx := context.Background()
	// Attempt 1
	//_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	//g.Expect(err).NotTo(HaveOccurred())
	//list := &sourcev1.GitRepositoryList{}
	//opts := client.MatchingLabels{
	//	"types.PartOfLabel": "appName",
	//}
	//err = k.List(ctx, list, &opts, client.InNamespace("msg.Namespace"))
	//g.Expect(err).ShouldNot(HaveOccurred())
	// Attempt 2
	//ctx := context.Background()
	//config := ctrl.GetConfigOrDie()
	//dynamic := dynamic.NewForConfigOrDie(config)
	////
	//namespace := "flux-system"
	//
	//resourceId := schema.GroupVersionResource{
	//	Group:    "",
	//	Version:  "v1",
	//	Resource: "bindings",
	//}
	//list, err := dynamic.Resource(resourceId).Namespace(namespace).
	//	List(ctx, metav1.ListOptions{})
	//
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	for _, item := range list.Items {
	//		fmt.Printf("%+v\n", item)
	//	}
	//}
	//Attempt 3 (Listing resources)
	config := ctrl.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error during Kubernetes client initialization, %s", err.Error())
		os.Exit(1)
	}

	apiGroups, apiResourceListArray, err := clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		fmt.Printf("Error during server resource discovery, %s", err.Error())
		os.Exit(1)
	}
	_ = apiGroups

	ctx := context.Background()
	dynamic := dynamic.NewForConfigOrDie(config)

	for _, apiResourceList := range apiResourceListArray {
		fmt.Printf("Group: %s\n", apiResourceList.TypeMeta.String())

		for _, apiResource := range apiResourceList.APIResources {
			fmt.Printf("\tResource => %s\n", apiResource.Name)

			canBeQueryied := false

			for _, verb := range apiResource.Verbs {
				if verb == "list" {
					canBeQueryied = true
				}
			}

			if !canBeQueryied {
				fmt.Printf("\t\tskiping as it cannot be queried \n")
				continue
			}

			if apiResource.Namespaced == false {
				fmt.Printf("\t\tskiping as it cannot be queried. It is not namespace based \n")
				continue
			}

			if apiResource.Name == "controllerrevisions" {
				fmt.Printf("")
			}

			groupInfo := strings.Split(apiResourceList.GroupVersion, "/")

			var group, version string
			if len(groupInfo) != 2 {
				group = ""
				version = groupInfo[0]
			} else {
				group = groupInfo[0]
				version = groupInfo[1]
			}

			resourceId := schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: apiResource.Name,
			}

			fmt.Printf("\t querying Group:%s Version:%s Resource:%s\n", apiResourceList.TypeMeta.APIVersion, apiResourceList.GroupVersion, apiResource.Name)

			list, err := dynamic.Resource(resourceId).Namespace("flux-system").List(ctx, metav1.ListOptions{})

			if err != nil {
				fmt.Println(err)
			} else {
				for _, item := range list.Items {
					fmt.Printf("\t\t%+v\n", item.GetName())
				}
			}
		}
		//if enableVerboseLogging {

		//fmt.Printf("Group: %s\n", apiResourceList.GroupVersion)
		////}
		//// rbac rules only look at API group names, not name & version
		//groupOnly := strings.Split(apiResourceList.GroupVersion, "/")[0]
		//// core API doesn't have a group "name". We set to "core" and replace at the end with a blank string in the rbac policy rule
		//if apiResourceList.GroupVersion == "v1" {
		//	groupOnly = "core"
		//}
		//
		//resourceList := make([]string, 0)
		//resourcesByVerb := make(map[string][]string)
		//
		//for _, apiResource := range apiResourceList.APIResources {
		//	//if enableVerboseLogging {
		//	fmt.Printf("Resource: %s - Verbs: %s\n",
		//		apiResource.Name,
		//		apiResource.Verbs.String())
		//	//}
		//	fmt.Printf("Verbs %s \n\n", apiResource.Verbs)
		//
		//	resourceList = append(resourceList, apiResource.Name)
		//
		//	canBeQueryied := false
		//	for _, verb := range apiResource.Verbs {
		//		if verb == "list" {
		//			canBeQueryied = true
		//		}
		//	}
		//
		//}
		//
		//for k := range resourcesByVerb {
		//	var sb strings.Builder
		//
		//	sb.WriteString(groupOnly)
		//	sb.WriteString("!")
		//	sb.WriteString(k)
		//	//resourcesByGroupAndVerb.Set(sb.String(), resourcesByVerb[k])
		//	fmt.Println("OUTPUT:", sb.String())
		//}
		fmt.Println("")
	}
}

func TestSingleCall(t *testing.T) {
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	dynamic := dynamic.NewForConfigOrDie(config)

	namespace := "flux-system"

	resourceId := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "controllerrevisions",
	}
	list, err := dynamic.Resource(resourceId).Namespace(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		fmt.Println(err)
	} else {
		for _, item := range list.Items {
			fmt.Printf("%+v\n", item)
		}
	}
}
