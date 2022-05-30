package clusters_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
)

func TestGetClusters(t *testing.T) {
	tests := []struct {
		name             string
		cs               []clusters.Cluster
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no clusters",
			expected: "No clusters found.\n",
		},
		{
			name: "clusters exist",
			cs: []clusters.Cluster{
				{
					Name: "cluster-a",
					Conditions: []clusters.Condition{
						{
							Type:    "Ready",
							Status:  "True",
							Message: "Cluster Found",
						},
					},
				},
				{
					Name: "cluster-b",
					Conditions: []clusters.Condition{
						{
							Type:    "Ready",
							Status:  "False",
							Message: "failed to get CAPI cluster",
						},
					},
				},
			},
			expected: "NAME\tSTATUS\tSTATUS_MESSAGE\ncluster-a\tReady\tCluster Found\ncluster-b\tNot Ready\tfailed to get CAPI cluster\n",
		},
		{
			name:             "error retrieving clusters",
			err:              fmt.Errorf("oops something went wrong"),
			expectedErrorStr: "unable to retrieve clusters from \"In-memory fake\": oops something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(tt.cs, "", tt.err)
			w := new(bytes.Buffer)
			err := clusters.GetClusters(c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetClusterByName(t *testing.T) {
	tests := []struct {
		name             string
		clusterName      string
		cs               []clusters.Cluster
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:     "no clusters",
			expected: "No clusters found.\n",
		},
		{
			name:        "cluster exist",
			clusterName: "cluster-a",
			cs: []clusters.Cluster{
				{
					Name: "cluster-a",
					Conditions: []clusters.Condition{
						{
							Type:   "Ready",
							Status: "True",
						},
					},
				},
			},
			expected: "NAME\tSTATUS\tSTATUS_MESSAGE\ncluster-a\tReady\t\n",
		},
		{
			name:        "not ready cluster",
			clusterName: "cluster-a",
			cs: []clusters.Cluster{
				{
					Name: "cluster-a",
					Conditions: []clusters.Condition{
						{
							Type:    "Ready",
							Status:  "False",
							Message: "failed to get CAPI cluster",
						},
					},
				},
			},
			expected: "NAME\tSTATUS\tSTATUS_MESSAGE\ncluster-a\tNot Ready\tfailed to get CAPI cluster\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(tt.cs, "", tt.err)
			w := new(bytes.Buffer)
			err := clusters.GetClusterByName(tt.clusterName, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestDeleteClusters(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:             "error returned",
			err:              errors.New("something went wrong"),
			expectedErrorStr: "unable to create pull request for cluster deletion: something went wrong",
		},
		{
			name:     "pull request created",
			result:   "https://github.com/org/repo/pull/1",
			expected: "Created pull request for clusters deletion: https://github.com/org/repo/pull/1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFakeClient(nil, tt.result, tt.err)
			w := new(bytes.Buffer)
			err := clusters.DeleteClusters(clusters.DeleteClustersParams{}, c, w)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

type FakeClient struct {
	cs  []clusters.Cluster
	s   string
	err error
}

func NewFakeClient(cs []clusters.Cluster, s string, err error) *FakeClient {
	return &FakeClient{
		cs:  cs,
		s:   s,
		err: err,
	}
}

func (c *FakeClient) Source() string {
	return "In-memory fake"
}

func (c *FakeClient) RetrieveClusters() ([]clusters.Cluster, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.cs, nil
}

func (c *FakeClient) GetClusterKubeconfig(name string) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}

func (c *FakeClient) DeleteClusters(params clusters.DeleteClustersParams) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.s, nil
}
