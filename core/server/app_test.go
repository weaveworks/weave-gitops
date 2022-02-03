package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateDeployKey(t *testing.T) {
	g := NewGomegaWithT(t)

	ctx := context.Background()

	h, rt := mockHttpClient()

	c, cleanup := makeGRPCServer(k8sEnv.Rest, h, t)
	defer cleanup()

	_, k, err := kube.NewKubeHTTPClientWithConfig(k8sEnv.Rest, "")
	g.Expect(err).NotTo(HaveOccurred())

	t.Run("[github] creates a deploy key secret on the cluster", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		repoURL, err := gitproviders.NewRepoURL("git@github.com/someorg/somerepo.git")
		g.Expect(err).NotTo(HaveOccurred())

		r := &pb.CreateDeployKeyRequest{
			SecretName: "my-secret",
			Namespace:  ns.Name,
			RepoUrl:    repoURL.String(),
			Provider:   pb.GitProvider_GitHub,
		}

		ctx = middleware.ContextWithGRPCAuth(ctx, "sometoken")

		rt.RoundTripStub = ghAPIRoundTripper(repoURL.Owner(), repoURL.RepositoryName())

		_, err = c.CreateDeployKey(ctx, r)
		g.Expect(err).NotTo(HaveOccurred())

		secret := &corev1.Secret{}
		g.Expect(k.Get(ctx, types.NamespacedName{Name: r.SecretName, Namespace: ns.Name}, secret)).To(Succeed())

		key := auth.ExtractPublicKey(secret)

		// Check for the key encryption algo name in the string.
		// We don't check for straight equality because the value should change randomly,
		// due to encryption.
		g.Expect(string(key)).To(ContainSubstring("ecdsa-sha2-nistp384 "))
	})

	t.Run("handles no secret name", func(t *testing.T) {
		ns := newNamespace(ctx, k, g)

		r := &pb.CreateDeployKeyRequest{
			SecretName: "",
			Namespace:  ns.Name,
			RepoUrl:    "git@github.com/someorg/somerepo.git",
			Provider:   pb.GitProvider_GitHub,
		}

		_, err := c.CreateDeployKey(ctx, r)
		g.Expect(err).To(HaveOccurred())
		status, ok := status.FromError(err)
		g.Expect(ok).To(BeTrue(), "could not get status from error")
		g.Expect(status.Code()).To(Equal(codes.InvalidArgument))

		secret := &corev1.Secret{}
		name := types.NamespacedName{Name: r.SecretName, Namespace: ns.Name}
		g.Expect(k.Get(ctx, name, secret)).NotTo(Succeed(), "cluster secret should not have been created")

	})
}

func ghAPIRoundTripper(owner, repoName string) func(r *http.Request) (*http.Response, error) {
	return func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case filepath.Join("/repos", owner, repoName):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{ "login": "github", "name": "github" }`)),
			}, nil
		case filepath.Join("/orgs", owner):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{ "login": "github", "name": "github" }`)),
			}, nil

		case filepath.Join("/repos", owner, repoName, "keys"):
			txt := `
			{
				"id": 1,
				"key": "SHA256:somecoolkey",
				"url": "https://api.github.com/repos/octocat/Hello-World/keys/1",
				"title": "wego-deploy-key",
				"verified": true,
				"created_at": "2014-12-10T15:53:42Z",
				"read_only": true
			  }		
`
			if r.Method == http.MethodGet {
				// GET returns a list of keys
				txt = fmt.Sprintf("[%s]", txt)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(txt)),
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusNotFound,
		}, errors.New("not found")
	}
}
