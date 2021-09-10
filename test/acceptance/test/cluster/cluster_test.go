package cluster

import (
	"testing"
)

func TestCluster(t *testing.T) {

	//showDbRecords("/var/folders/24/xpk2zmbj2552mxs2nhzsq8t00000gn/T/db-directory2553224797")
	//dbDirectory, err := ioutil.TempDir("", "db-directory")
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = CreateClusterDB(dbDirectory)
	//if err != nil {
	//	panic(err)
	//}
	//
	//clusterPool2 := NewClusterPool2()
	//
	//clusterPool2.GenerateClusters2(dbDirectory)

	//c := make(chan string,2)
	//go func() {
	//	time.Sleep(time.Second*3)
	//	c<-"one"
	//}()
	//
	//fmt.Println("VALUE",<-c)

	// Fails creating a cluster the first time

	//c := make(chan string,1)
	//go func(){
	//    time.Sleep(time.Second*5)
	//    c<- "hello"
	//}()
	//fmt.Println("c",<-c)
	//
	//
	//time.Sleep(time.Second)

	//clusterPool := NewClusterPool()
	//clusterPool.Generate()
	//
	//var wg sync.WaitGroup
	//
	//wg.Add(1)
	//go func() {
	//   defer wg.Done()
	//   fmt.Println("waiting to get cluster name....")
	//   fmt.Println("got cluster name",clusterPool.GetNextClusterName())
	//}()
	//
	//wg.Wait()
	//
	//fmt.Println("Waiting 10 seconds")
	//time.Sleep(time.Second*10)
	//
	//clusterPool.End()

}
