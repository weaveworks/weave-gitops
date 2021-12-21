package profiles

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/gogo/protobuf/jsonpb"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
)

const (
	wegoServiceName = "wego-app"
	getProfilesPath = "/v1/profiles"
)

type GetOptions struct {
	Namespace string
	ClientSet kubernetes.Interface
	Writer    io.Writer
	Port      string
}

func GetProfiles(ctx context.Context, opts GetOptions) error {
	profiles, err := getProfiles(ctx, opts)
	if err != nil {
		return err
	}

	printProfiles(profiles, opts.Writer)

	return nil
}

func getProfiles(ctx context.Context, opts GetOptions) (*pb.GetProfilesResponse, error) {
	resp, err := kubernetesDoRequest(ctx, opts.Namespace, wegoServiceName, opts.Port, getProfilesPath, opts.ClientSet)
	if err != nil {
		return nil, err
	}

	profiles := &pb.GetProfilesResponse{}
	err = jsonpb.UnmarshalString(string(resp), profiles)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return profiles, nil
}

func printProfiles(profiles *pb.GetProfilesResponse, w io.Writer) {
	fmt.Fprintf(w, "NAME\tDESCRIPTION\tAVAILABLE_VERSIONS\n")

	if profiles.Profiles != nil && len(profiles.Profiles) > 0 {
		for _, p := range profiles.Profiles {
			fmt.Fprintf(w, "%s\t%s\t%v", p.Name, p.Description, strings.Join(p.AvailableVersions, ","))
			fmt.Fprintln(w, "")
		}
	}
}

func kubernetesDoRequest(ctx context.Context, namespace, serviceName, servicePort, path string, clientset kubernetes.Interface) ([]byte, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	data, err := clientset.CoreV1().Services(namespace).ProxyGet("http", serviceName, servicePort, u.String(), nil).DoRaw(ctx)
	if err != nil {
		if se, ok := err.(*errors.StatusError); ok {
			return nil, fmt.Errorf("failed to make GET request to service %s/%s path %q status code: %d", namespace, serviceName, path, int(se.Status().Code))
		}

		return nil, fmt.Errorf("failed to make GET request to service %s/%s path %q: %w", namespace, serviceName, path, err)
	}

	return data, nil
}
