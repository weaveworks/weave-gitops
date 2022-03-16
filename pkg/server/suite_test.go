package server

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server")
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

var s *grpc.Server
var apps pb.ApplicationsServer
var appsClient pb.ApplicationsClient
var conn *grpc.ClientConn
var err error
var k8sClient client.Client
var testClustername = "test-cluster"
var scheme *apiruntime.Scheme
var k kube.Kube
var ghAuthClient *authfakes.FakeGithubAuthClient
var gitProvider *gitprovidersfakes.FakeGitProvider
var glAuthClient *authfakes.FakeGitlabAuthClient
var configGit *gitfakes.FakeGit
var env *testutils.K8sTestEnv
var fakeFactory *servicesfakes.FakeFactory
var jwtClient auth.JWTClient

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

var _ = BeforeSuite(func() {
	scheme = kube.CreateScheme()

	env, err = testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})
	Expect(err).NotTo(HaveOccurred())

	k8sClient = env.Client
})

var _ = AfterSuite(func() {
	env.Stop()
})

var secretKey string

var _ = BeforeEach(func() {
	lis = bufconn.Listen(bufSize)
	s = grpc.NewServer()

	rand.Seed(time.Now().UnixNano())
	secretKey = rand.String(20)

	k = &kube.KubeHTTP{
		Client:      k8sClient,
		ClusterName: testClustername,
		DynClient:   env.DynClient,
		RestMapper:  k8sClient.RESTMapper(),
	}

	gitProvider = &gitprovidersfakes.FakeGitProvider{}
	gitProvider.GetDefaultBranchReturns("main", nil)

	fakeFactory = &servicesfakes.FakeFactory{}
	configGit = &gitfakes.FakeGit{}

	fakeFactory.GetGitClientsReturns(configGit, gitProvider, nil)

	ghAuthClient = &authfakes.FakeGithubAuthClient{}
	glAuthClient = &authfakes.FakeGitlabAuthClient{}
	jwtClient = auth.NewJwtClient(secretKey)
	fakeClientGetter := kubefakes.NewFakeClientGetter(k8sClient)
	fakeKubeGetter := kubefakes.NewFakeKubeGetter(k)

	cfg := ApplicationsConfig{
		Factory:          fakeFactory,
		JwtClient:        jwtClient,
		GithubAuthClient: ghAuthClient,
		GitlabAuthClient: glAuthClient,
		ClusterConfig:    kube.ClusterConfig{},
	}
	apps = NewApplicationsServer(&cfg,
		WithClientGetter(fakeClientGetter), WithKubeGetter(fakeKubeGetter))
	pb.RegisterApplicationsServer(s, apps)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	ctx := context.Background()
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))

	Expect(err).NotTo(HaveOccurred())

	appsClient = pb.NewApplicationsClient(conn)
})

var _ = AfterEach(func() {
	conn.Close()
})
