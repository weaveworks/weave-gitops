package cluster

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Cluster struct {
	Name           string
	Context        string
	KubeConfigPath string
}

func NewCluster(name string, context string, kubeConfigPath string) *Cluster {
	return &Cluster{
		Name:           name,
		Context:        context,
		KubeConfigPath: kubeConfigPath,
	}
}

func (c *Cluster) CleanUp() {
	c.delete()
	c.deleteKubeConfigFile()
}

func (c *Cluster) delete() {
	cmd := fmt.Sprintf("kind delete cluster --name %s --kubeconfig %s", c.Name, c.KubeConfigPath)
	command := exec.Command("sh", "-c", cmd)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func (c *Cluster) deleteKubeConfigFile() {
	err := os.RemoveAll(c.KubeConfigPath)
	Expect(err).ShouldNot(HaveOccurred())
}

// TODO: Start generating unit tests for ClusterPool
// TODO: Remove last kubeconfigfile and last cluster after error or on end
// TODO: Hability to pass in the name of the cluster you want
// TODO: Generalize paths of kubeconfig, etc.

func CreateKindCluster(ctx context.Context, rootKubeConfigFilesPath string) (*Cluster, error) {
	supportedProviders := "kind"
	supportedK8SVersions := "1.19.1, 1.20.2, 1.21.1"

	provider, found := os.LookupEnv("CLUSTER_PROVIDER")
	if !found {
		provider = "kind"
	}

	k8sVersion, found := os.LookupEnv("K8S_VERSION")
	if !found {
		k8sVersion = "1.20.2"
	}

	if !strings.Contains(supportedProviders, provider) {
		log.Errorf("Cluster provider %s is not supported for testing", provider)
		return nil, errors.New("Unsupported provider")
	}

	if !strings.Contains(supportedK8SVersions, k8sVersion) {
		log.Errorf("Kubernetes version %s is not supported for testing", k8sVersion)
		return nil, errors.New("Unsupported kubernetes version")
	}

	var cluster *Cluster

	if provider == "kind" {
		clusterName := RandString(6)
		kubeConfigFile := "kube-config-" + clusterName
		kubeConfigPath := filepath.Join(string(rootKubeConfigFilesPath), kubeConfigFile)
		log.Infof("Creating a kind cluster %s", clusterName)

		c := fmt.Sprintf("kind create cluster --name=%s --kubeconfig %s --image=%s --config=configs/kind-config.yaml --wait 5m", clusterName, kubeConfigPath, "kindest/node:v"+k8sVersion)
		cmd := exec.CommandContext(ctx, "sh", "-c", c)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Infof("Failed to create kind cluster")
			//log.Fatal(err)
			return nil, err
		}
		cluster = NewCluster(clusterName, "kind-"+clusterName, kubeConfigPath)

	}

	return cluster, nil
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return StringWithCharset(length, charset)
}

type ClusterStatus string

const (
	ClusterCreating  ClusterStatus = "CREATING"
	ClusterRequested ClusterStatus = "REQUESTED"
	ClusterCreated   ClusterStatus = "CREATED"
	ClusterBeingUsed ClusterStatus = "BEINGUSED"
	ClusterDeleted   ClusterStatus = "DELETED"
)

const CLUSTER_DB = "db"
const CLUSTER_TABLE = "clusters"

type Cluster2 struct {
	Name            string
	Context         string
	KubeConfigPath  string
	Status          ClusterStatus
	ErrorOnCreation error
}

func (c *Cluster2) CleanUp() {
	c.delete()
	c.deleteKubeConfigFile()
}

func (c *Cluster2) delete() {
	cmd := fmt.Sprintf("kind delete cluster --name %s --kubeconfig %s", c.Name, c.KubeConfigPath)
	command := exec.Command("sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		fmt.Printf("ERROR deleting cluster %s\n", err)
	}
}

func (c *Cluster2) deleteKubeConfigFile() {
	err := os.RemoveAll(c.KubeConfigPath)
	if err != nil {
		fmt.Printf("ERROR deleting kubeconfig %s\n", err)
	}
}

func NewCluster2(name string, context string, kubeConfigPath string, status ClusterStatus) *Cluster2 {
	return &Cluster2{
		Name:           name,
		Context:        context,
		KubeConfigPath: kubeConfigPath,
		Status:         status,
	}
}

func convertCluster2ToBytes(cluster Cluster2) []byte {
	bts, _ := json.Marshal(cluster)
	return bts
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func CreateClusterDB(dbPath string) error {
	db, err := bolt.Open(filepath.Join(dbPath, CLUSTER_DB), 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(CLUSTER_TABLE))
		return err
	})
}

func CreateClusterRecord2(dbPath string, cluster Cluster) error {

	db, err := bolt.Open(filepath.Join(dbPath, CLUSTER_DB), 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLUSTER_TABLE))
		id, _ := b.NextSequence()
		clusterID := itob(int(id))

		c2 := Cluster2{
			Name:           cluster.Name,
			KubeConfigPath: cluster.KubeConfigPath,
			Context:        cluster.Context,
			Status:         ClusterCreated,
		}

		bts := convertCluster2ToBytes(c2)

		return b.Put(clusterID, bts)
	})

	return err
}

func RequestClusterCreation(dbPath []byte) error {

	db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLUSTER_TABLE))
		id, _ := b.NextSequence()
		clusterID := itob(int(id))
		return b.Put(clusterID, convertCluster2ToBytes(Cluster2{Status: ClusterRequested}))
	})

}

func FindCreatedClusterAndAssignItToSomeRecord(dbPath []byte) ([]byte, Cluster2, error) {

	var cc Cluster2
	var kClusterID []byte
	for {
		cc = Cluster2{Name: ""}
		kClusterID = make([]byte, 0)

		db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
		if err != nil {
			return nil, cc, fmt.Errorf("error opening db in get cluster %w", err)
		}

		err = db.Batch(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(CLUSTER_TABLE))
			return b.ForEach(func(clusterID, v []byte) error {
				c := Cluster2{}
				err := json.Unmarshal(v, &c)
				if err != nil {
					return fmt.Errorf("error on unmarshal on iteration0 %w", err)
				}

				if c.Status == ClusterCreated && len(kClusterID) == 0 {
					kClusterID = append(make([]byte, 0, len(clusterID)), clusterID...)
					err := json.Unmarshal(v, &cc)
					if err != nil {
						return fmt.Errorf("error on unmarshal on iteration1 %w", err)
					}
				}
				return nil
			})
		})
		if err != nil {
			return nil, cc, err
		}

		if len(kClusterID) != 0 {
			err = db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(CLUSTER_TABLE))
				cc.Status = ClusterBeingUsed
				bts, err := json.Marshal(cc)
				if err != nil {
					return err
				}
				return b.Put(kClusterID, bts)
			})
			if err != nil {
				return nil, cc, err
			}
		}

		err = db.Close()
		if err != nil {
			return nil, cc, err
		}

		if len(kClusterID) != 0 {
			break
		}

		time.Sleep(time.Second * 5)
	}

	return kClusterID, cc, nil
}

func UpdateClusterToDeleted(dbPath []byte, clusterID []byte, cluster Cluster2) error {
	db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLUSTER_TABLE))
		cluster.Status = ClusterDeleted
		bts, err := json.Marshal(&cluster)
		if err != nil {
			return err
		}
		return b.Put(clusterID, bts)
	})
}

type ClusterPool2 struct {
	errOnGenerate             []error
	listenToRequestedClusters bool
	sync.RWMutex
}

func (c *ClusterPool2) Errors() []error {
	c.Lock()
	defer c.Unlock()
	return c.errOnGenerate
}

func NewClusterPool2() *ClusterPool2 {
	return &ClusterPool2{listenToRequestedClusters: true, errOnGenerate: make([]error, 0)}
}

func (c *ClusterPool2) End() {
	c.Lock()
	c.listenToRequestedClusters = false
	c.Unlock()

}

func (c *ClusterPool2) AppendError(err error) {
	c.Lock()
	c.errOnGenerate = append(c.errOnGenerate, err)
	c.Unlock()
	gexec.Kill()
}

func (c *ClusterPool2) IsListeningToRequestedClusters() bool {
	c.Lock()
	defer c.Unlock()
	return c.listenToRequestedClusters
}

// CreateClusterOnRequest
func (c *ClusterPool2) CreateClusterOnRequest(ctx context.Context, dbPath string) {
	// iterate over all register until find REQUESTED

	for c.IsListeningToRequestedClusters() {

		db, err := bolt.Open(filepath.Join(dbPath, CLUSTER_DB), 0755, &bolt.Options{})
		if err != nil {
			c.AppendError(fmt.Errorf("error opening db %w", err))
		}

		var kClusterID []byte
		err = db.Batch(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(CLUSTER_TABLE))
			return b.ForEach(func(clusterID, v []byte) error {
				c := Cluster2{}
				err := json.Unmarshal(v, &c)
				if err != nil {
					return fmt.Errorf("error on unmarshal on iteration0 %w", err)
				}

				if c.Status == ClusterRequested && kClusterID == nil {
					kClusterID = clusterID
					c.Status = ClusterCreating
					if err = b.Put(kClusterID, convertCluster2ToBytes(c)); err != nil {
						return fmt.Errorf("error on unmarshal on iteration3 %w", err)
					}
				}
				return nil
			})
		})
		if err != nil {
			c.AppendError(fmt.Errorf("error on db batch %w", err))
		}
		err = db.Close()
		if err != nil {
			c.AppendError(fmt.Errorf("error closing db connection %w", err))
		}

		if kClusterID != nil {

			kindCluster, err := CreateKindCluster(ctx, dbPath)
			if err != nil {
				c.AppendError(fmt.Errorf("error creating kind cluster %w", err))
			}

			if kindCluster != nil {
				db, err = bolt.Open(filepath.Join(dbPath, CLUSTER_DB), 0755, &bolt.Options{})
				if err != nil {
					c.AppendError(fmt.Errorf("error opening db %w", err))
				}
				err = db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(CLUSTER_TABLE))
					cluster := NewCluster2(kindCluster.Name, kindCluster.Context, kindCluster.KubeConfigPath, ClusterCreated)
					bts, err := json.Marshal(cluster)
					if err != nil {
						return fmt.Errorf("error unmarshalling cluster %w", err)
					}
					err = b.Put(kClusterID, bts)
					if err != nil {
						return fmt.Errorf("error updating cluster record %w", err)
					}
					return err
				})
				if err != nil {
					c.AppendError(fmt.Errorf("error on db update %w", err))
				}
				err = db.Close()
				if err != nil {
					c.AppendError(fmt.Errorf("error closing db connection %w", err))
				}
			}
		}

		time.Sleep(time.Second * 10)
	}
}

func (c *ClusterPool2) GenerateClusters2(dbPath string, clusterCount int) {

	ctx := context.Background()

	clusters := make(chan *Cluster, clusterCount)
	done := make(chan bool, 1)
	go func() {
		for cluster := range clusters {
			if cluster != nil {
				err := CreateClusterRecord2(dbPath, *cluster)
				if err != nil {
					c.AppendError(fmt.Errorf("error creating record %w", err))
				}
			}
		}
		done <- true
	}()

	var wg sync.WaitGroup

	for i := 0; i < clusterCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kindCluster, err := CreateKindCluster(ctx, dbPath)
			if err != nil {
				c.AppendError(fmt.Errorf("error creating kind cluster %w", err))
			}
			clusters <- kindCluster
		}()
	}

	wg.Wait()

	close(clusters)

	<-done

}
