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

type ClusterPool struct {
	cluster        chan *Cluster
	lastCluster    *Cluster
	end            bool
	err            error
	kubeConfigRoot string
}

// TODO: Start generating unit tests for ClusterPool
// TODO: Remove last kubeconfigfile and last cluster after error or on end
// TODO: Hability to pass in the name of the cluster you want
// TODO: Generalize paths of kubeconfig, etc.

func NewClusterPool(size int) *ClusterPool {
	return &ClusterPool{cluster: make(chan *Cluster, size)}
}

func (c *ClusterPool) GetNextCluster() *Cluster {
	return <-c.cluster
}

//func (c *ClusterPool) Generate() error {
//
//	kubeConfigRoot, err := ioutil.TempDir("", "kube-config")
//	if err != nil {
//		return err
//	}
//	c.kubeConfigRoot = kubeConfigRoot
//	fmt.Println("Creating kube config files on ", kubeConfigRoot)
//
//	go func() {
//		for !c.end {
//			cluster, err := CreateKindCluster(kubeConfigRoot)
//			if err != nil {
//				c.err = err
//				break
//			}
//			c.lastCluster = cluster
//			c.cluster <- cluster
//		}
//	}()
//
//	return nil
//}

func (c *ClusterPool) Error() error {
	return c.err
}

func (c *ClusterPool) End() {
	c.lastCluster.CleanUp()
	c.end = true
}

func CreateFakeCluster(ind int64) (string, error) {
	return fmt.Sprintf("fakeClusterName%d", ind), nil
}

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

		pwd, _ := os.Getwd()
		fmt.Println("Creating cluster... ", pwd)
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
		fmt.Println("Cluster DONE")
		cluster = NewCluster(clusterName, "kind-"+clusterName, kubeConfigPath)

	}

	return cluster, nil
}

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	fmt.Println("CMD", cmd)
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
		fmt.Println("")
		cc = Cluster2{Name: ""}
		kClusterID = make([]byte, 0)

		db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
		if err != nil {
			return nil, cc, fmt.Errorf("error opening db in get cluster %w", err)
		}

		fmt.Println("BEFORE db.Batch")
		err = db.Batch(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(CLUSTER_TABLE))
			return b.ForEach(func(clusterID, v []byte) error {
				c := Cluster2{}
				err := json.Unmarshal(v, &c)
				if err != nil {
					return fmt.Errorf("error on unmarshal on iteration0 %w", err)
				}

				if c.Status == ClusterCreated && len(kClusterID) == 0 {
					fmt.Printf("\tCLUSTER-ID-FOUND clusterID[%v] clusterID==nil[%v] \n", clusterID, clusterID == nil)
					kClusterID = append(make([]byte, 0, len(clusterID)), clusterID...)
					fmt.Printf("\tCLUSTER-ID-FOUND222 clusterID[%v] clusterID==nil[%v] \n", kClusterID, kClusterID == nil)
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
		fmt.Printf("AFTER db.Batch clusterID[%v] clusterID==nil[%v] \n", kClusterID, kClusterID == nil)

		if len(kClusterID) != 0 {
			fmt.Printf("INSIDE update to status=ClusterBeingUsed clusterID[%v] clusterID==nil[%v] \n", kClusterID, kClusterID == nil)
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
		fmt.Printf("AFTER if kClusterID != nil { clusterID[%v] clusterID==nil[%v] \n", kClusterID, kClusterID == nil)

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

func GetCluster(dbPath []byte, clusterID []byte) (*Cluster2, error) {

	db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0666, &bolt.Options{})
	if err != nil {
		return nil, fmt.Errorf("error opening db in get cluster %w", err)
	}
	defer db.Close()

	c := Cluster2{}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLUSTER_TABLE))
		bts := b.Get(clusterID)
		return json.Unmarshal(bts, &c)
	})
	if err != nil {
		return nil, fmt.Errorf("error on getting cluster %w", err)
	}

	return &c, nil
}

func UpdateClusterToBeingUsed(dbPath []byte, clusterID []byte, cluster *Cluster2) error {

	db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLUSTER_TABLE))
		cluster.Status = ClusterBeingUsed
		bts, err := json.Marshal(*cluster)
		if err != nil {
			return err
		}
		return b.Put(clusterID, bts)
	})

	if err != nil {
		return fmt.Errorf("error updating cluster to BeingUsed %w", err)
	}

	return nil
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
		fmt.Printf("clusterID==nil234 [%v] \n", clusterID == nil)
		fmt.Printf("clusterID234 [%v] \n", clusterID)
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
		fmt.Printf("CreateClusterOnRequest %v \n", kClusterID == nil)

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
					//clusterID[%v] clusterID==nil[%v] \n",kClusterID,kClusterID==nil)
					fmt.Sprintf("kClusterID==nil [%v]\n", kClusterID == nil)
					fmt.Sprintf("kClusterID456 [%v]\n", kClusterID)
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

	// get ID then create kind cluster then update status to CREATED
}

func (c *ClusterPool2) GenerateClusters2(dbPath string, clusterCount int) {

	ctx := context.Background()

	clusters := make(chan *Cluster, clusterCount)
	done := make(chan bool, 1)
	go func() {
		for cluster := range clusters {
			err := CreateClusterRecord2(dbPath, *cluster)
			if err != nil {
				c.AppendError(fmt.Errorf("error creating record %w", err))
			}
		}
		done <- true
	}()

	var wg sync.WaitGroup

	//fmt.Printf("Creating %d %d %d clusters\n", config.GinkgoConfig.ParallelTotal, config.GinkgoConfig.ParallelTotal*2, config.GinkgoConfig.ParallelNode)
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

//func (c *ClusterPool2) GenerateClusters(dbPath []byte) error {
//
//	// Write errors to a state to be able to see them at the end
//	// and also as a way of breaking the loop early
//	// we might want to write error state to all records so we can break waiting no clusters to be created
//
//	// BREAK THIS LOOP BY SOME VARIABLE
//
//	c.Lock()
//	keepGoing := c.end
//	c.Unlock()
//
//	for !keepGoing {
//		kindCluster, err := CreateKindCluster(dbPath)
//		if err != nil {
//			return err
//		}
//
//		clusterAssigned := false
//		for !clusterAssigned {
//
//			db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
//			if err != nil {
//				return err
//			}
//
//			err = db.Batch(func(tx *bolt.Tx) error {
//				b := tx.Bucket([]byte(CLUSTER_TABLE))
//				cc := Cluster2{}
//				var clusterWaitingID []byte
//				clustersBeingUsed := 0
//				clusterWaiting := 0
//				b.ForEach(func(clusterID, v []byte) error {
//					err := json.Unmarshal(v, &cc)
//					if err != nil {
//						return fmt.Errorf("error on unmarshal on iteration %w", err)
//					}
//					if cc.Status == ClusterBeingUsed {
//						clustersBeingUsed++
//					}
//					if cc.Status == ClusterWaiting {
//						clusterWaiting++
//						clusterWaitingID = clusterID
//					}
//
//					return nil
//				})
//
//				fmt.Println("CLUSTERS-WAITING ", clusterWaiting)
//
//				if clustersBeingUsed < config.GinkgoConfig.ParallelTotal*2 && len(clusterWaitingID) != 0 {
//					// assign kind cluster to cluster
//					// find one cluster, set to creating, and start creating the cluster
//					d := NewCluster2(kindCluster.Name, kindCluster.Context, kindCluster.KubeConfigPath, ClusterCreated)
//					bts, err := json.Marshal(*d)
//					if err != nil {
//						panic(err)
//						return fmt.Errorf("error marshalling on assingment %w", err)
//					}
//					if err = b.Put(clusterWaitingID, bts); err != nil {
//						panic(err)
//						return fmt.Errorf("error updating cluster on assigment %w", err)
//					}
//					fmt.Printf("WRITTING KIND CONFIG TO CLUSTER RECORD %s\n", bts)
//					clusterAssigned = true
//				} else {
//					c.LatestCluster = kindCluster
//				}
//
//				return nil
//
//			})
//			if err != nil {
//				return err
//			}
//
//			if err = db.Close(); err != nil {
//				return err
//			}
//
//			// Give some time before checking again for cluster status changes
//			time.Sleep(time.Second * 5)
//		}
//	}
//
//	fmt.Println("DONE creating clusters")
//
//	return nil
//}

//db.View(func(tx *bolt.Tx) error {
//	b := tx.Bucket([]byte("MyBucket"))
//	v := b.Get([]byte("answer"))
//	fmt.Printf("The answer is: %s\n", v)
//	return nil
//})

func createDB() error {
	db, err := bolt.Open("my.db", 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	//var clusterID2 []byte
	//err = db.Update(func(tx *bolt.Tx) error {
	//
	//	tx.CreateBucketIfNotExists([]byte(CLUSTER_DB))
	//
	//	b := tx.Bucket([]byte(CLUSTER_TABLE))
	//
	//	id, _ := b.NextSequence()
	//	clusterID := int(id)
	//
	//	cluster := NewCluster2(clusterID,"name1","ctx1","path")
	//
	//	clusterID2 = itob(cluster.ID)
	//
	//	return b.Put(clusterID2,convertCluster2ToBytes(cluster))
	//})
	//if err != nil {
	//	return err
	//}

	//err = db.Batch(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte(CLUSTER_TABLE))
	//	bts := b.Get(clusterID2)
	//	c := &Cluster2{}
	//	err := json.Unmarshal(bts,c)
	//	if err != nil {
	//		return err
	//	}
	//	fmt.Println("CName",c.Name)
	//	fmt.Println("CContext",c.Context)
	//	fmt.Println("CKubeConfigPath",c.KubeConfigPath)
	//	return nil
	//})
	//if err != nil {
	//	return err
	//}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte(CLUSTER_TABLE))

			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("key1=%s, value=%s\n", k, v)
			}

			return nil
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte(CLUSTER_TABLE))

			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("key0=%s, value=%s\n", k, v)
			}

			return nil
		})
	}()

	wg.Wait()

	return nil

	//db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte(CLUSTER_TABLE))
	//
	//	c := b.Cursor()
	//
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		fmt.Printf("key=%s, value=%s\n", k, v)
	//	}
	//
	//	return nil
	//})

}

func showDbRecords(dbPath string) {

	db, err := bolt.Open(filepath.Join(string(dbPath), CLUSTER_DB), 0755, &bolt.Options{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Checking DB records")
	err = db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CLUSTER_TABLE))
		return b.ForEach(func(clusterID, v []byte) error {
			fmt.Printf("Record %v %s \n", clusterID, v)
			return nil
		})
	})
	if err != nil {
		panic(err)
	}
}
