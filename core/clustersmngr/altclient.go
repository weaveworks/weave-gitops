package clustersmngr

import (
	"context"
	"fmt"
	"sync"

	gitopsv1alpha1 "github.com/weaveworks/cluster-controller/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewAltClient() (Client, error) {

	return &altClient{}, nil
}

type altClient struct {
	Client
}

func (ac *altClient) ClusteredList(ctx context.Context, clist ClusteredObjectList, namespaced bool, opts ...client.ListOption) error {
	// build a client from values extracted from ctx or read from cache
	fmt.Println("doing other client")
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("building in-cluster config: %w", err)
	}

	user := auth.Principal(ctx)

	fmt.Println(user.Token())

	// id="jordan@weave.works" groups=[wge-test-org-jp:team-a wge-test-org-jp:team-b]

	cfg.BearerToken = user.Token()
	// Set this to zero to ensure the in-cluster config does not read from file for the token.
	cfg.BearerTokenFile = ""

	scheme, err := kube.CreateScheme()
	if err != nil {
		return fmt.Errorf("creating scheme: %w", err)
	}

	cc, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	clusterList := &gitopsv1alpha1.GitopsClusterList{}

	// get a list of clusters
	if err := cc.List(ctx, clusterList); err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}

	fmt.Printf("len: %v\n", len(clusterList.Items))

	clients := map[string]client.Client{}

	for _, clust := range clusterList.Items {

		s := &corev1.Secret{}

		t := types.NamespacedName{
			Name:      clust.Spec.SecretRef.Name,
			Namespace: clust.GetNamespace(),
		}

		if err := cc.Get(ctx, t, s); err != nil {
			return fmt.Errorf("getting secret: %w", err)
		}

		cfgBytes := s.Data["value"]

		restCfg, err := clientcmd.RESTConfigFromKubeConfig(cfgBytes)
		if err != nil {
			return fmt.Errorf("creating rest config from secret: %w", err)
		}

		clusterClient, err := client.New(restCfg, client.Options{})
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		clients[clust.GetName()] = clusterClient
	}

	// for each cluster
	// get a list of namespaces
	for clusterName, cli := range clients {
		nsList := &corev1.NamespaceList{}

		if err := cli.List(ctx, nsList); err != nil {
			return fmt.Errorf("listing namespaces: %w", err)
		}

		checker := nsaccess.NewChecker(nsaccess.DefautltWegoAppRules)

		cs, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return fmt.Errorf("getting clientset: %w", err)
		}

		// do selfsubjectacccessreviews and store the available namespaces in the cached client
		filteredNs, err := checker.FilterAccessibleNamespaces(ctx, cs.AuthorizationV1(), nsList.Items)
		if err != nil {
			return fmt.Errorf("filtering namespaces: %w", err)
		}

		var (
			errs = ClusteredListError{}
			wg   = sync.WaitGroup{}
		)

		for _, n := range filteredNs {
			// for each namespace
			// list objects
			wg.Add(1)

			go func(clusterName string, nsName string, c client.Client, optsWithNamespace ...client.ListOption) {
				defer wg.Done()

				list := clist.NewList()

				ctx, cancel := context.WithTimeout(ctx, clientTimeout)
				defer cancel()

				if err := c.List(ctx, list, optsWithNamespace...); err != nil {
					errs.Add(ListError{Cluster: clusterName, Namespace: nsName, Err: err})
				}

				clist.AddObjectList(clusterName, list)
			}(clusterName, n.Name, cc, client.InNamespace(n.Name))
		}

		wg.Wait()

	}

	// cache the resulting client (?)
	// Q: should we do this on login?
	// Q: can we take the user to some special loading page while we build their client?
	// - imagining a special button they click on the login screen that navigates to a page to trigger this work.
	// - show progress
	// - bust the client cache
	// - guard against retries that will saturate the api server
	// - mutex to lock the go routine that gets cleaned up
	// - endpoint to hit to generate client
	// - endpoint to give client generation status; SSE?

	// our cached client might get a 403 if the user's namespace access has been revoked,
	// but we can just ignore the 403 and update the available namespaces in the stored client.

	return nil
}
