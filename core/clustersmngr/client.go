package clustersmngr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is wrapper to controller-runtime/client adding multi clusters context.
// it contains the list of clusters and namespaces the user has access to allowing
// cross cluster/namespace querying
type Client interface {
	// Get retrieves an obj for the given object key.
	Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error
	// List retrieves list of objects for a given namespace and list options.
	List(ctx context.Context, cluster string, list client.ObjectList, opts ...client.ListOption) error

	// Create saves the object obj.
	Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error
	// Delete deletes the given obj
	Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error
	// Update updates the given obj.
	Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error
	// Patch patches the given obj
	Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error

	// ClusteredList loops through the list of clusters and namespaces the client has access and
	// queries the list of objects for each of them in parallel.
	// This method supports pagination with a caveat, the client.Limit passed will be multiplied
	// by the number of clusters and namespaces, we decided to do this to avoid the complex coordination
	// that would be required to make sure the number of items returned match the limit passed.
	ClusteredList(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error

	// ClientsPool returns the clients pool.
	ClientsPool() ClientsPool

	// Scoped returns a client that is scoped to a single cluster
	Scoped(cluster string) (client.Client, error)
}

type clustersClient struct {
	pool       ClientsPool
	namespaces map[string][]v1.Namespace
}

type ListError struct {
	Cluster   string
	Namespace string
	Err       error
}

func (le ListError) Error() string {
	return fmt.Sprintf("Failed to list resource on cluster=%q namespace=%q err=%q", le.Cluster, le.Namespace, le.Err)
}

type ClusteredListError struct {
	Errors []ListError
}

func (cle *ClusteredListError) Add(err ListError) {
	cle.Errors = append(cle.Errors, err)
}

func (cle ClusteredListError) Error() string {
	var errs []string
	for _, e := range cle.Errors {
		errs = append(errs, e.Error())
	}

	return strings.Join(errs, "; ")
}

func NewClient(clientsPool ClientsPool, namespaces map[string][]v1.Namespace) Client {
	return &clustersClient{
		pool:       clientsPool,
		namespaces: namespaces,
	}
}

func (c *clustersClient) ClientsPool() ClientsPool {
	return c.pool
}

func (c *clustersClient) Get(ctx context.Context, cluster string, key client.ObjectKey, obj client.Object) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Get(ctx, key, obj)
}

func (c *clustersClient) List(ctx context.Context, cluster string, list client.ObjectList, opts ...client.ListOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.List(ctx, list, opts...)
}

func (c *clustersClient) ClusteredList(ctx context.Context, clist ClusteredObjectList, opts ...client.ListOption) error {
	paginationInfo := &PaginationInfo{}

	continueToken := extractContinueToken(opts...)
	if continueToken != "" {
		if err := decodeFromBase64(paginationInfo, continueToken); err != nil {
			return fmt.Errorf("failed decoding pagination info: %w", err)
		}
	}

	var (
		errs = ClusteredListError{}
		wg   = sync.WaitGroup{}
	)

	for clusterName, cc := range c.pool.Clients() {
		for _, ns := range c.namespaces[clusterName] {
			nsContinueToken := paginationInfo.Get(clusterName, ns.Name)

			// a prior request has been made so this one comes with a previous token,
			// but if the namespace token is empty we ignore it because all items have been returned.
			if continueToken != "" && nsContinueToken == "" {
				continue
			}

			listOpts := append(opts, client.Continue(nsContinueToken))
			listOpts = append(listOpts, client.InNamespace(ns.Name))

			wg.Add(1)

			go func(clusterName string, nsName string, c client.Client, optsWithNamespace ...client.ListOption) {
				defer wg.Done()

				list := clist.NewList()

				if err := c.List(ctx, list, optsWithNamespace...); err != nil {
					errs.Add(ListError{Cluster: clusterName, Namespace: nsName, Err: err})
				}

				paginationInfo.Set(clusterName, nsName, list.GetContinue())

				clist.AddObjectList(clusterName, list)
			}(clusterName, ns.Name, cc, listOpts...)
		}
	}

	wg.Wait()

	continueToken, err := encodeToBase64(paginationInfo)
	if err != nil {
		return fmt.Errorf("failed encoding pagination info: %w", err)
	}

	clist.SetContinue(continueToken)

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func extractContinueToken(opts ...client.ListOption) string {
	for _, o := range opts {
		switch v := o.(type) {
		case client.Continue:
			return string(v)
		default:
			continue
		}
	}

	return ""
}

func (c *clustersClient) Create(ctx context.Context, cluster string, obj client.Object, opts ...client.CreateOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Create(ctx, obj, opts...)
}

func (c *clustersClient) Delete(ctx context.Context, cluster string, obj client.Object, opts ...client.DeleteOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Delete(ctx, obj, opts...)
}

func (c *clustersClient) Update(ctx context.Context, cluster string, obj client.Object, opts ...client.UpdateOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Update(ctx, obj, opts...)
}

func (c *clustersClient) Patch(ctx context.Context, cluster string, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return err
	}

	return client.Patch(ctx, obj, patch, opts...)
}

func (c clustersClient) Scoped(cluster string) (client.Client, error) {
	client, err := c.pool.Client(cluster)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// ClusteredObjectList represents the returns of the lists of all clusters and namespaces user could query
type ClusteredObjectList interface {
	// NewList is a factory that returns a new concrete list being queried
	NewList() client.ObjectList
	// AddObjectList adds a result list of objects to the lists map
	AddObjectList(cluster string, list client.ObjectList)
	// Lists returns the map of lists from all clusters
	Lists() map[string][]client.ObjectList
	// GetContinue returns the continue token used for pagination
	GetContinue() string
	// SetContinue sets the continue token used for pagination
	SetContinue(continueToken string)
}

type ClusteredList struct {
	sync.Mutex

	listFactory   func() client.ObjectList
	lists         map[string][]client.ObjectList
	continueToken string
}

func NewClusteredList(listFactory func() client.ObjectList) ClusteredObjectList {
	return &ClusteredList{
		listFactory: listFactory,
		lists:       make(map[string][]client.ObjectList),
	}
}

func (cl *ClusteredList) NewList() client.ObjectList {
	return cl.listFactory()
}

func (cl *ClusteredList) AddObjectList(cluster string, list client.ObjectList) {
	cl.Lock()
	defer cl.Unlock()

	cl.lists[cluster] = append(cl.lists[cluster], list)
}

func (cl *ClusteredList) Lists() map[string][]client.ObjectList {
	cl.Lock()
	defer cl.Unlock()

	return cl.lists
}

func (cl *ClusteredList) GetContinue() string {
	return cl.continueToken
}

func (cl *ClusteredList) SetContinue(continueToken string) {
	cl.continueToken = continueToken
}

type PaginationInfo struct {
	sync.Mutex
	ContinueTokens map[string]map[string]string
}

func (pi *PaginationInfo) Set(cluster string, namespace string, token string) {
	pi.Lock()
	defer pi.Unlock()

	if pi.ContinueTokens == nil {
		pi.ContinueTokens = make(map[string]map[string]string)
	}

	if pi.ContinueTokens[cluster] == nil {
		pi.ContinueTokens[cluster] = make(map[string]string)
	}

	pi.ContinueTokens[cluster][namespace] = token
}

func (pi *PaginationInfo) Get(cluster string, namespace string) string {
	pi.Lock()
	defer pi.Unlock()

	if pi.ContinueTokens == nil {
		return ""
	}

	if pi.ContinueTokens[cluster] == nil {
		return ""
	}

	if val, ok := pi.ContinueTokens[cluster][namespace]; ok {
		return val
	}

	return ""
}

func decodeFromBase64(v interface{}, enc string) error {
	return json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(enc))).Decode(v)
}

func encodeToBase64(v interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	err := json.NewEncoder(encoder).Encode(v)
	if err != nil {
		return "", err
	}

	encoder.Close()

	return buf.String(), nil
}
