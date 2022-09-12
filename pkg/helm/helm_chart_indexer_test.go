package helm

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"k8s.io/apimachinery/pkg/types"
)

// ObjectReference points to a resource.
type ObjectReference struct {
	Kind       string
	APIVersion string
	Name       string
	Namespace  string
}

type Chart struct {
	Name    string
	Version string
}

// HelmChartIndexer indexs details of Helm charts that have been seen in Helm
// repositories.
type HelmChartIndexer struct {
	CacheDB *sql.DB
}

// AddChart inserts a new chart into helm_charts table.
func (i *HelmChartIndexer) AddChart(ctx context.Context, name, version string, clusterRef types.NamespacedName, repoRef ObjectReference) error {
	sqlStatement := `
INSERT INTO helm_charts (name, version,
	repo_kind, repo_api_version, repo_name, repo_namespace,
	cluster_name, cluster_namespace)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := i.CacheDB.ExecContext(
		ctx,
		sqlStatement, name, version,
		repoRef.Kind, repoRef.APIVersion, repoRef.Name, repoRef.Namespace,
		clusterRef.Name, clusterRef.Namespace)

	return err
}

func (i *HelmChartIndexer) Count(ctx context.Context) (int64, error) {
	rows, err := i.CacheDB.QueryContext(ctx, "SELECT COUNT(*) FROM helm_charts")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return 0, err
		}
		count += n
	}

	return count, nil
}

// ListChartsByCluster returns a list of charts filtered by cluster.
func (i *HelmChartIndexer) ListChartsByCluster(ctx context.Context, clusterRef types.NamespacedName) ([]Chart, error) {
	sqlStatement := `
SELECT name, version FROM helm_charts 
WHERE cluster_name = $1 AND cluster_namespace = $2`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var chart Chart
		if err := rows.Scan(&chart.Name, &chart.Version); err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

// ListChartsByRepositoryAndCluster returns a list of charts filtered by helm repository and cluster.
func (i *HelmChartIndexer) ListChartsByRepositoryAndCluster(ctx context.Context, repoRef ObjectReference, clusterRef types.NamespacedName) ([]Chart, error) {
	sqlStatement := `
SELECT name, version FROM helm_charts 
WHERE repo_kind = $1 AND repo_api_version = $2 AND repo_name = $3 AND repo_namespace = $4
AND cluster_name = $5 AND cluster_namespace = $6`

	rows, err := i.CacheDB.QueryContext(ctx, sqlStatement, repoRef.Kind, repoRef.APIVersion, repoRef.Name, repoRef.Namespace, clusterRef.Name, clusterRef.Namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []Chart
	for rows.Next() {
		var chart Chart
		if err := rows.Scan(&chart.Name, &chart.Version); err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

func TestHelmChartIndex(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	count, err := indexer.Count(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d, want 1", count)
	}
}

func TestListChartsByCluster(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1",
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	charts, err := indexer.ListChartsByCluster(context.TODO(), nsn("cluster1", "clusters"))
	if err != nil {
		t.Fatal(err)
	}
	if len(charts) != 2 {
		t.Fatalf("got %d, want 2", len(charts))
	}
}

func TestListChartsByRepositoryAndCluster(t *testing.T) {
	db := createDB(t)
	indexer := HelmChartIndexer{
		CacheDB: db,
	}

	if err := indexer.AddChart(context.TODO(), "redis", "1.0.1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1",
		nsn("cluster1", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}
	if err := indexer.AddChart(context.TODO(), "nginx", "1.0.1",
		nsn("cluster2", "clusters"),
		objref("HelmRepository", "", "bitnami-charts", "team-ns")); err != nil {
		t.Fatal(err)
	}

	charts, err := indexer.ListChartsByRepositoryAndCluster(context.TODO(),
		objref("HelmRepository", "", "bitnami-charts", "team-ns"),
		nsn("cluster1", "clusters"))
	if err != nil {
		t.Fatal(err)
	}
	if len(charts) != 2 {
		t.Fatalf("got %d, want 2", len(charts))
	}
}

func nsn(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func objref(kind, apiVersion, name, namespace string) ObjectReference {
	return ObjectReference{
		Kind:       kind,
		APIVersion: apiVersion,
		Name:       name,
		Namespace:  namespace,
	}
}

func applySchema(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS helm_charts (
	name text, version text,
	repo_kind text, repo_api_version text, repo_name text, repo_namespace text,
	cluster_name text, cluster_namespace text);
	`)
	return err
}

func createDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})
	if err := applySchema(db); err != nil {
		t.Fatal(err)
	}

	return db
}
