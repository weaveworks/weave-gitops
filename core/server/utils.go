package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/weaveworks/weave-gitops/core/server/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	var opts client.MatchingLabels
	if appName != "" {
		opts = client.MatchingLabels{
			types.PartOfLabel: appName,
		}
	}

	return opts
}

func printOutAllResources(namespace string, labelsMap map[string]string) {
	config := ctrl.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error during Kubernetes client initialization, %s", err.Error())
		os.Exit(1)
	}

	_, apiResourceListArray, err := clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		fmt.Printf("Error during server resource discovery, %s", err.Error())
		os.Exit(1)
	}

	ctx := context.Background()
	dynamic := dynamic.NewForConfigOrDie(config)

	for _, apiResourceList := range apiResourceListArray {
		fmt.Printf("Group: %s\n", apiResourceList.GroupVersion)

		for _, apiResource := range apiResourceList.APIResources {
			fmt.Printf("\tResource => %s\n", apiResource.Name)

			canBeQueryied := false

			for _, verb := range apiResource.Verbs {
				if verb == "list" {
					canBeQueryied = true
				}
			}

			if !canBeQueryied {
				fmt.Printf("\t\tskiping as it cannot be queried. There is not list verb \n")
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

			fmt.Printf("\t\t querying Group:%s Version:%s Resource:%s\n", apiResourceList.TypeMeta.APIVersion, apiResourceList.GroupVersion, apiResource.Name)

			labelSelector := metav1.LabelSelector{
				MatchLabels: labelsMap,
			}

			listOptions := metav1.ListOptions{
				LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
			}

			list, err := dynamic.Resource(resourceId).Namespace(namespace).List(ctx, listOptions)

			if err != nil {
				fmt.Println(err)
			} else {
				for _, item := range list.Items {
					fmt.Printf("\t\t\t%+v\n", item.GetName())
				}
			}
		}
	}
}
