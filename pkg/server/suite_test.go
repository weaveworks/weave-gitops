package server

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/apputils/apputilsfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	corev1 "k8s.io/api/core/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
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
var testEnv *envtest.Environment
var testClustername = "test-cluster"
var cfg *rest.Config
var scheme *apiruntime.Scheme
var k kube.Kube
var k8sManager ctrl.Manager
var ghAuthClient *authfakes.FakeGithubAuthClient
var gp *gitprovidersfakes.FakeGitProvider
var appGit *gitfakes.FakeGit
var configGit *gitfakes.FakeGit

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

var _ = BeforeSuite(func() {
	done := make(chan interface{})
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../manifests/crds",
			"../../tools/testcrds",
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	scheme = kube.CreateScheme()

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		ClientDisableCacheFor: []client.Object{
			&wego.Application{},
			&corev1.Namespace{},
			&corev1.Secret{},
		},
		Scheme: scheme,
	})
	Expect(err).ToNot(HaveOccurred())
	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	go func() {
		Eventually(done, 60).Should(BeClosed())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())
	close(done)
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
		DynClient:   dynamic.NewForConfigOrDie(k8sManager.GetConfig()),
		RestMapper:  k8sClient.RESTMapper(),
	}

	osysClient := osys.New()

	gp = &gitprovidersfakes.FakeGitProvider{}
	gp.GetDefaultBranchStub = func(s string) (string, error) {
		return "main", nil
	}

	appFactory := &apputilsfakes.FakeAppFactory{}

	appGit = &gitfakes.FakeGit{}
	configGit = &gitfakes.FakeGit{}

	appFactory.GetAppServiceForAddStub = func(c context.Context, params apputils.AddServiceParams) (app.AppService, error) {
		return &app.App{
			Context:     context.Background(),
			AppGit:      appGit,
			ConfigGit:   configGit,
			Flux:        flux.New(osysClient, &testutils.LocalFluxRunner{Runner: &runner.CLIRunner{}}),
			Kube:        k,
			Logger:      &loggerfakes.FakeLogger{},
			Osys:        osysClient,
			GitProvider: gp,
		}, nil
	}

	appFactory.GetKubeServiceStub = func() (kube.Kube, error) {
		return k, nil
	}

	ghAuthClient = &authfakes.FakeGithubAuthClient{}

	cfg := ApplicationsConfig{
		AppFactory:       appFactory,
		JwtClient:        auth.NewJwtClient(secretKey),
		KubeClient:       k8sClient,
		GithubAuthClient: ghAuthClient,
	}
	apps = NewApplicationsServer(&cfg)
	pb.RegisterApplicationsServer(s, apps)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	ctx := context.Background()
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	Expect(err).NotTo(HaveOccurred())

	appsClient = pb.NewApplicationsClient(conn)
})

var _ = AfterEach(func() {
	conn.Close()
})
