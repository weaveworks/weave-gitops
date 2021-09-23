package cluster

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	log "github.com/sirupsen/logrus"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type KindCluster struct {
	Name           string
	Context        string
	KubeConfigPath string
}

func NewCluster(name string, context string, kubeConfigPath string) *KindCluster {
	return &KindCluster{
		Name:           name,
		Context:        context,
		KubeConfigPath: kubeConfigPath,
	}
}

func (c *KindCluster) CleanUp() {
	c.delete()
	c.deleteKubeConfigFile()
}

func (c *KindCluster) delete() {
	cmd := fmt.Sprintf("kind delete cluster --name %s --kubeconfig %s", c.Name, c.KubeConfigPath)
	command := exec.Command("sh", "-c", cmd)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func (c *KindCluster) deleteKubeConfigFile() {
	err := os.RemoveAll(c.KubeConfigPath)
	Expect(err).ShouldNot(HaveOccurred())
}

func CreateKindCluster(ctx context.Context, rootKubeConfigFilesPath string) (*KindCluster, error) {
	supportedK8SVersions := "1.19.1, 1.20.2, 1.21.1"

	k8sVersion, found := os.LookupEnv("K8S_VERSION")
	if !found {
		k8sVersion = "1.20.2"
	}

	if !strings.Contains(supportedK8SVersions, k8sVersion) {
		log.Errorf("Kubernetes version %s is not supported for testing", k8sVersion)
		return nil, errors.New("unsupported kubernetes version")
	}

	var cluster *KindCluster

	clusterName := RandString(30)
	kubeConfigFile := "kube-config-" + clusterName
	kubeConfigPath := filepath.Join(rootKubeConfigFilesPath, kubeConfigFile)

	log.Infof("Creating a kind cluster %s", clusterName)

	c := fmt.Sprintf("kind create cluster --name=%s --kubeconfig %s --image=%s --config=configs/kind-config.yaml --wait 5m", clusterName, kubeConfigPath, "kindest/node:v"+k8sVersion)
	cmd := exec.CommandContext(ctx, "sh", "-c", c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Info("Failed to create kind cluster")
		log.Fatal(err)

		return nil, err
	}

	cluster = NewCluster(clusterName, fmt.Sprintf("kind-%s", clusterName), kubeConfigPath)

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

	err := command.Run()
	if err != nil {
		fmt.Printf("error deleting cluster %s\n", err)
	}
}

func (c *Cluster2) deleteKubeConfigFile() {
	err := os.RemoveAll(c.KubeConfigPath)
	if err != nil {
		fmt.Printf("error deleting kubeconfig %s\n", err)
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

func CreateClusterRecord2(dbPath string, cluster KindCluster) error {
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

type ClusterPool struct {
	listenToRequestedClusters bool
	sync.RWMutex
}

func NewClusterPool() *ClusterPool {
	return &ClusterPool{listenToRequestedClusters: true}
}

func (c *ClusterPool) End() {
	c.Lock()
	c.listenToRequestedClusters = false
	c.Unlock()
}

func (c *ClusterPool) IsListeningToRequestedClusters() bool {
	c.Lock()
	defer c.Unlock()

	return c.listenToRequestedClusters
}

// CreateClusterOnRequest
func (c *ClusterPool) CreateClusterOnRequest(ctx context.Context, dbPath string) {

	clusterIDs := make(chan []byte, 1)

	go func() {
		for c.IsListeningToRequestedClusters() {
			db, err := bolt.Open(filepath.Join(dbPath, CLUSTER_DB), 0755, &bolt.Options{})
			if err != nil {
				log.Fatal(fmt.Errorf("error opening db %w", err))
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
						kClusterID = append(make([]byte, 0, len(clusterID)), clusterID...)
						c.Status = ClusterCreating
						if err = b.Put(kClusterID, convertCluster2ToBytes(c)); err != nil {
							return fmt.Errorf("error on unmarshal on iteration3 %w", err)
						}
					}
					return nil
				})
			})

			if err != nil {
				log.Fatal(fmt.Errorf("error on db batch %w", err))
			}

			err = db.Close()

			if err != nil {
				log.Fatal(fmt.Errorf("error closing db connection %w", err))
			}

			if kClusterID != nil {
				clusterIDs <- kClusterID
			}

			time.Sleep(time.Second * 10)
		}

		close(clusterIDs)
	}()

	var wg sync.WaitGroup
	for cID := range clusterIDs {
		wg.Add(1)
		go func(cID []byte) {
			defer wg.Done()
			kindCluster, err := CreateKindCluster(ctx, dbPath)
			if err != nil {
				log.Fatal(fmt.Errorf("error creating kind cluster %w", err))
			}

			db, err := bolt.Open(filepath.Join(dbPath, CLUSTER_DB), 0755, &bolt.Options{})
			if err != nil {
				log.Fatal(fmt.Errorf("error opening db %w", err))
			}

			err = db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(CLUSTER_TABLE))
				cluster := NewCluster2(kindCluster.Name, kindCluster.Context, kindCluster.KubeConfigPath, ClusterCreated)
				bts, err := json.Marshal(cluster)
				if err != nil {
					return fmt.Errorf("error unmarshalling cluster %w", err)
				}
				err = b.Put(cID, bts)
				if err != nil {
					return fmt.Errorf("error updating cluster record %w", err)
				}
				return err
			})

			if err != nil {
				log.Fatal(fmt.Errorf("error on db update %w", err))
			}

			err = db.Close()

			if err != nil {
				log.Fatal(fmt.Errorf("error closing db connection %w", err))
			}

		}(cID)
	}

	wg.Wait()
}

func (c *ClusterPool) GenerateClusters(dbPath string, clusterCount int) {
	ctx := context.Background()

	clusters := make(chan *KindCluster, clusterCount)
	done := make(chan bool, 1)

	go func() {
		for cluster := range clusters {
			if cluster != nil {
				err := CreateClusterRecord2(dbPath, *cluster)
				if err != nil {
					log.Fatal(fmt.Errorf("error creating record %w", err))
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
				log.Fatal(fmt.Errorf("error creating kind cluster %w", err))
			}
			clusters <- kindCluster
		}()
	}

	wg.Wait()

	close(clusters)

	<-done
}
