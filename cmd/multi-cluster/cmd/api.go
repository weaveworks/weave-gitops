package multi_cluster

import (
	"context"
	"fmt"
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UserCtxKey = iota
	ClustersClientsPoolCtxKey
)

type User struct {
	Email string
	Team  string
}

// simulates OIDC autenticate that potentially adds the user to context
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Injecting authenticated user
		ctx := context.WithValue(r.Context(), UserCtxKey, User{Email: "luiz.filho@weave.gitops", Team: "developers"})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// creates clusters client for provided user in the context
func clustersClientsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Making sure user is present in the request and has been authenticated
		user, ok := r.Context().Value(UserCtxKey).(User)

		if !ok && user.Email == "" && user.Team == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, `{"error": "there is no authorized user"}`) // TODO: use proper structs for this error message
			return
		}

		// TODO: we should probably cache everything down here after we check the user is present,
		// to avoid creating a bunch of clients in every single call
		clustersFetcher, err := NewConfigMapClustersFetcher()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "failed getting management cluster client: %s"}`, err) // TODO: use proper structs for this error message
			return
		}

		clusters, err := clustersFetcher.Fetch(r.Context()) // TODO: cache it
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `{"error": "failed fetching clusters list: %w"}`, err) // TODO: use proper structs for this error message
			return
		}

		// TODO: cache it
		clientsPool := NewClustersClientsPool()
		for _, c := range clusters {
			if err := clientsPool.Add(user, c); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error": "failed adding cluster client to the pool: %s"}`, err) // TODO: use proper structs for this error message
				return
			}
		}

		ctx := context.WithValue(r.Context(), ClustersClientsPoolCtxKey, clientsPool)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func final(w http.ResponseWriter, r *http.Request) {
	clientsPool, ok := r.Context().Value(ClustersClientsPoolCtxKey).(*clientsPool)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error": "failed fetching clusters clients"}`) // TODO: use proper structs for this error message

		return
	}

	for name, c := range clientsPool.Clients() {
		podList := &corev1.PodList{}

		err := c.List(context.Background(), podList, client.InNamespace("developers"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "failed to list pods for cluster=%s"}`, name) // TODO: use proper structs for this error message

			return
		}

		fmt.Fprintln(w, podList)
	}

	w.Write([]byte("OK"))
}

func NewServer() error {
	mux := http.NewServeMux()

	finalHandler := http.HandlerFunc(final)

	mux.Handle("/", authMiddleware(clustersClientsMiddleware(finalHandler)))

	log.Println("Listening on :3000...")

	return http.ListenAndServe(":3000", mux)
}
