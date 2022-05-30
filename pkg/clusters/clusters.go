package clusters

import (
	"fmt"
	"io"
)

// ClustersRetriever defines the interface that adapters
// need to implement in order to return an array of clusters.
type ClustersRetriever interface {
	Source() string
	RetrieveClusters() ([]Cluster, error)
	GetClusterKubeconfig(string) (string, error)
	DeleteClusters(DeleteClustersParams) (string, error)
}

type Cluster struct {
	Name       string      `json:"name"`
	Conditions []Condition `json:"conditions"`
}

type Condition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GetClusters uses a ClustersRetriever adapter to show
// a list of clusters to the console.
func GetClusters(r ClustersRetriever, w io.Writer) error {
	cs, err := r.RetrieveClusters()
	if err != nil {
		return fmt.Errorf("unable to retrieve clusters from %q: %w", r.Source(), err)
	}

	if len(cs) > 0 {
		fmt.Fprintf(w, "NAME\tSTATUS\tSTATUS_MESSAGE\n")

		for _, c := range cs {
			printCluster(c, w)
		}

		return nil
	}

	fmt.Fprintf(w, "No clusters found.\n")

	return nil
}

// GetClusterByName uses a ClustersRetriever adapter to show
// a cluster to the console given its name.
func GetClusterByName(name string, r ClustersRetriever, w io.Writer) error {
	cs, err := r.RetrieveClusters()
	if err != nil {
		return fmt.Errorf("unable to retrieve clusters from %q: %w", r.Source(), err)
	}

	if len(cs) > 0 {
		fmt.Fprintf(w, "NAME\tSTATUS\tSTATUS_MESSAGE\n")

		for _, c := range cs {
			if c.Name == name {
				printCluster(c, w)
			}
		}

		return nil
	}

	fmt.Fprintf(w, "No clusters found.\n")

	return nil
}

func GetClusterKubeconfig(name string, r ClustersRetriever, w io.Writer) error {
	k, err := r.GetClusterKubeconfig(name)
	if err != nil {
		return fmt.Errorf("unable to retrieve cluster %q from %q: %w", name, r.Source(), err)
	}

	fmt.Fprint(w, k)

	return nil
}

func DeleteClusters(params DeleteClustersParams, r ClustersRetriever, w io.Writer) error {
	pr, err := r.DeleteClusters(params)
	if err != nil {
		return fmt.Errorf("unable to create pull request for cluster deletion: %w", err)
	}

	fmt.Fprintf(w, "Created pull request for clusters deletion: %s\n", pr)

	return nil
}

type DeleteClustersParams struct {
	GitProviderToken string
	RepositoryURL    string
	HeadBranch       string
	BaseBranch       string
	Title            string
	Description      string
	ClustersNames    []string
	CommitMessage    string
}

func printCluster(c Cluster, w io.Writer) {
	var status, message string

	for _, condition := range c.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				status = condition.Type
			} else {
				status = "Not Ready"
			}

			message = condition.Message
		}
	}

	fmt.Fprintf(w, "%s\t%s\t%s", c.Name, status, message)
	fmt.Fprintln(w, "")
}
