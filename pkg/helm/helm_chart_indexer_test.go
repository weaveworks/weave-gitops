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

// HelmChartIndexer indexs details of Helm charts that have been seen in Helm
// repositories.
type HelmChartIndexer struct {
	CacheDB *sql.DB
}

func (i *HelmChartIndexer) AddChart(ctx context.Context, name, version string, clusterRef types.NamespacedName, repoRef ObjectReference) error {

	return nil
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
CREATE TABLE helm_charts (
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
