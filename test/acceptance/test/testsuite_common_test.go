// +build !unittest

package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/test/acceptance/test/cluster"
)

func TestAcceptance(t *testing.T) {

	if testing.Short() {
		t.Skip("Skip User Acceptance Tests")
	}

	RegisterFailHandler(Fail)
	gomega.RegisterFailHandler(GomegaFail)
	RunSpecs(t, "Weave GitOps User Acceptance Tests")
}

var clusterPool2 *cluster.ClusterPool2

//var syncCluster2 *cluster.Cluster2

// TODO: crear todos los kind clusters al mismo tiempo al inicio
// TODO: despues solo crear uno cuando se desocupe uno
// TODO: delete temporary root folder
// TODO: solo crear un nuevo cluster cuando se elimine uno
// asi me evito de tener tanta logica en el método de generar
// claro!!! de esta lo unico que tendria que hacer es generar
// los cluster paralelos que voy a querer as inicio y ya dejo
// que lo que las creaciones despues de eliminar sean las que "generen mas"
// Ya está!!! solo tengo que crear lo doble de los nodos al inicio(N) para ya tener
// listo N cluster cuando el primero termino y ya con la logic de crear al eliminar
// siempre tendria N disponibles

var globalCtx context.Context
var globalCancel func()

var _ = SynchronizedBeforeSuite(func() []byte {
	//clusterPool = cluster.NewClusterPool()
	//err := clusterPool.Generate()
	//Expect(err).NotTo(HaveOccurred())

	// go routine to generate kind clusters based on the nodes number N
	//   it will have N clusters running at all times
	//   cluster with deleted=false and cluster = waiting is available for selection
	// it will iterate records in order
	//   if all records have cluster = created then stop
	// save error field to save if an error occurred at the moment of creation attempt

	dbDirectory, err := ioutil.TempDir("", "db-directory")
	Expect(err).NotTo(HaveOccurred())

	fmt.Println("TEMP-DIRECTORY", dbDirectory)

	err = cluster.CreateClusterDB(dbDirectory)
	Expect(err).NotTo(HaveOccurred())

	clusterPool2 = cluster.NewClusterPool2()

	clusterPool2.GenerateClusters2(dbDirectory, config.GinkgoConfig.ParallelTotal)
	go clusterPool2.GenerateClusters2(dbDirectory, 1)

	globalCtx, globalCancel = context.WithCancel(context.Background())

	go clusterPool2.CreateClusterOnRequest(globalCtx, dbDirectory)

	return []byte(dbDirectory)
}, func(dbDirectory []byte) {

	fmt.Println("Running Node ", config.GinkgoConfig.ParallelNode)

	globalDbDirectory = dbDirectory

	SetDefaultEventuallyTimeout(EVENTUALLY_DEFAULT_TIME_OUT)
	DEFAULT_SSH_KEY_PATH = os.Getenv("HOME") + "/.ssh/id_rsa"
	GITHUB_ORG = os.Getenv("GITHUB_ORG")
	WEGO_BIN_PATH = os.Getenv("WEGO_BIN_PATH")
	if WEGO_BIN_PATH == "" {
		WEGO_BIN_PATH = "/usr/local/bin/wego"
	}
	log.Infof("WEGO Binary Path: %s", WEGO_BIN_PATH)

	//var err error
	//syncCluster, err = cluster.CreateKindCluster(string(kubeConfigRoot))
	//Expect(err).NotTo(HaveOccurred())
	//syncCluster = clusterPool.GetNextCluster()

	//  func(clusterPoolSyncFile []byte) {
	//    calculate randomID
	//    write randomID
	//    waitUntil the record pointing to randomID has a cluster on it
	//    if error then fail with Expected
	//    createClusterReferences syncCluster based on record
})

func GomegaFail(message string, callerSkip ...int) {
	if webDriver != nil {
		filepath := takeScreenshot()
		fmt.Printf("Failure screenshot is saved in file %s\n", filepath)
	}
	ginkgo.Fail(message, callerSkip...)
}

var _ = SynchronizedAfterSuite(func() {
	//err := cluster.UpdateClusterToDeleted(gloablDbDirectory,globalclusterID,syncCluster)
	//Expect(err).NotTo(HaveOccurred())
	//syncCluster.CleanUp()
}, func() {
	//globalCancel()
	clusterPool2.End()
	cmd := "kind delete clusters --all"
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		fmt.Printf("Error deleting ramaining clusters %s\n", err)
	}
	err = os.RemoveAll(string(globalDbDirectory))
	if err != nil {
		fmt.Printf("Error deleting root folder %s\n", err)
	}
	errors := clusterPool2.Errors()
	if len(errors) > 0 {
		for _, err := range clusterPool2.Errors() {
			fmt.Println("error", err)
		}
	}
})

//var _ = BeforeSuite(func() {
//
//})
