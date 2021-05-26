package server_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	client "github.com/weaveworks/weave-gitops/pkg/client/v1"
	"github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
	pb "github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
	"github.com/weaveworks/weave-gitops/pkg/server"
)

const port = ":50051"
const sessionID = "123-abc"

func createServer(t *testing.T) *http.Server {

	handler := server.NewServer()
	server := http.Server{Addr: port, Handler: handler}

	go func() {
		if err := server.ListenAndServe(); err != nil && err.Error() != http.ErrServerClosed.Error() {
			fmt.Println(err.Error())
		}
	}()

	return &server
}

func getUrl() *url.URL {
	u, _ := url.Parse(fmt.Sprintf("http://localhost%s", port))
	return u
}

func createClient(t *testing.T, c *http.Client) gitops.GitOps {
	return client.NewClient(getUrl().String(), c)
}

func createUnauthenticatedClient(t *testing.T) gitops.GitOps {
	return createClient(t, http.DefaultClient)
}

func createAuthenticatedClient(t *testing.T) gitops.GitOps {
	clientWithAuth := server.CreateAuthenticatedClient(t, getUrl(), "my-user-id-123")

	return createClient(t, clientWithAuth)
}

func Test_ListApplications(t *testing.T) {
	client := createAuthenticatedClient(t)
	s := createServer(t)

	ctx := context.Background()

	defer s.Shutdown(ctx)

	res, err := client.ListApplications(ctx, &pb.ListApplicationsReq{})

	if err != nil {
		t.Fatal(err)
	}

	if len(res.Applications) == 0 {
		t.Errorf("Expected more than zero apps, got %v", len(res.Applications))
	}
}

func Test_ListApplications_Unauthenticated(t *testing.T) {
	client := createUnauthenticatedClient(t)
	s := createServer(t)

	ctx := context.Background()

	defer s.Shutdown(ctx)

	_, err := client.ListApplications(ctx, &pb.ListApplicationsReq{})

	if err == nil {
		t.Fatal(errors.New("expected a 401 from the server"))
	}
}
