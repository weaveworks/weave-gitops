//go:build !unittest
// +build !unittest

package server_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	corev1 "k8s.io/api/core/v1"
)

const (
	bufSize           = 1024 * 1024
	gitlabTokenEnvVar = "GITLAB_TOKEN"
	gitlabOrgEnvVar   = "GITLAB_ORG"

	githubTokenEnvVar = "GITHUB_TOKEN"
	githubOrgEnvVar   = "GITHUB_ORG"
)

var (
	lis         *bufconn.Listener
	env         *testutils.K8sTestEnv
	conn        *grpc.ClientConn
	s           *grpc.Server
	err         error
	clusterName = "test-cluster"
	gitlabToken string
	gitlabOrg   string
	githubToken string
	githubOrg   string
)

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

var stop func()

func TestServerIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server Integration")
}

var _ = BeforeSuite(func() {
	gitlabToken = getEnvVar(gitlabTokenEnvVar)
	gitlabOrg = getEnvVar(gitlabOrgEnvVar)

	githubToken = getEnvVar(githubTokenEnvVar)
	githubOrg = getEnvVar(githubOrgEnvVar)

	ctx := context.Background()
	env, err = testutils.StartK8sTestEnvironment([]string{
		"../../../manifests/crds",
		"../../../tools/testcrds",
	})
	Expect(err).NotTo(HaveOccurred())

	fluxNs := &corev1.Namespace{}
	fluxNs.Name = "flux-system"

	Expect(env.Client.Create(ctx, fluxNs)).To(Succeed())

	stop = env.Stop
	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
	fluxClient.SetupBin()

	factory := services.NewServerFactory(fluxClient, &loggerfakes.FakeLogger{}, env.Rest, clusterName)
	Expect(err).NotTo(HaveOccurred())

	cfg := &server.ApplicationsConfig{
		Factory:          factory,
		Logger:           zapr.NewLogger(zap.NewNop()),
		JwtClient:        auth.NewJwtClient("somekey"),
		GithubAuthClient: auth.NewGithubAuthClient(http.DefaultClient),
		FetcherFactory:   server.NewDefaultFetcherFactory(),
		ClusterConfig: server.ClusterConfig{
			DefaultConfig: env.Rest,
			ClusterName:   clusterName,
		},
	}

	s = grpc.NewServer()
	apps := server.NewApplicationsServer(cfg)
	pb.RegisterApplicationsServer(s, apps)

	go func() {
		if err := s.Serve(lis); err != nil {
			fmt.Println(err.Error())
		}
	}()

	lis = bufconn.Listen(bufSize)

	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	stop()
	conn.Close()
	s.Stop()
})

func getEnvVar(envVar string) string {
	value := os.Getenv(envVar)
	ExpectWithOffset(1, value).NotTo(BeEmpty(), fmt.Sprintf("Please ensure %s environment variable is set", envVar))

	return value
}
