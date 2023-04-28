package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"

	. "github.com/onsi/gomega"

	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/minio/minio-go/v7"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockGet struct {
	client.Client
}

var _ = client.Client(&mockGet{})

func (m *mockGet) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	switch obj := obj.(type) {
	case *corev1.Secret:
		obj.Data = map[string][]byte{
			"accesskey": []byte("abcd"),
			"secretkey": []byte("1234"),
		}
	case *sourcev1.Bucket:
		obj.Spec.Endpoint = "endpoint:9000"
		obj.Spec.Insecure = false
	}

	return nil
}

func TestGetBucketConnectionInfo(t *testing.T) {
	g := NewGomegaWithT(t)

	type args struct {
		ctx         context.Context
		clusterName string
		ns          string
		cli         client.Client
	}

	tests := []struct {
		name    string
		args    args
		want    *bucketConnectionInfo
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx:         context.TODO(),
				clusterName: "Default",
				ns:          "default",
				cli:         &mockGet{},
			},
			want: &bucketConnectionInfo{
				accessKey:      "abcd",
				secretKey:      "1234",
				bucketEndpoint: "endpoint:9000",
				bucketInsecure: false,
			},
			wantErr: false,
		},
		{
			name: "test",
			args: args{
				ctx:         context.TODO(),
				clusterName: "my-session/run-session",
				ns:          "default",
				cli:         &mockGet{},
			},
			want: &bucketConnectionInfo{
				accessKey:      "abcd",
				secretKey:      "1234",
				bucketEndpoint: "run-session-bucket.my-session.svc:9000",
				bucketInsecure: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		info, err := getBucketConnectionInfo(tt.args.ctx, tt.args.clusterName, tt.args.ns, tt.args.cli)
		g.Expect(err != nil).To(Equal(tt.wantErr))
		g.Expect(info).To(Equal(tt.want))
	}
}

type mockS3Reader struct {
}

var _ = s3Reader(&mockS3Reader{})

func (m *mockS3Reader) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo)

	go func() {
		defer close(ch)

		switch bucketName {
		case "test":
			ch <- minio.ObjectInfo{
				Key:  "test",
				Size: 4,
			}
		case "error":
			ch <- minio.ObjectInfo{
				Key: "error",
			}
		}
	}()

	return ch
}

var timeFixture = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

func (m *mockS3Reader) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	switch objectName {
	case "test":
		o := &pb.LogEntry{
			Timestamp: timeFixture.Format(time.RFC3339),
			Message:   "test",
			Level:     "info",
			Source:    "gitops-run-client",
		}
		b, err := json.Marshal(o)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(string(b))), nil
	case "error":
		return nil, fmt.Errorf("error")
	}

	return nil, fmt.Errorf("not found")
}

func TestGitOpsRunLogs(t *testing.T) {
	g := NewGomegaWithT(t)

	type args struct {
		ctx        context.Context
		sessionID  string
		nextToken  string
		minio      s3Reader
		bucketName string
	}

	tests := []struct {
		name      string
		args      args
		want      []*pb.LogEntry
		wantToken string
		wantErr   bool
	}{
		{
			name: "test",
			args: args{
				ctx:        context.TODO(),
				sessionID:  "test",
				nextToken:  "test",
				minio:      &mockS3Reader{},
				bucketName: "test",
			},
			want: []*pb.LogEntry{{
				Level:     "info",
				Message:   "test",
				Source:    "gitops-run-client",
				Timestamp: timeFixture.Format(time.RFC3339),
			}},
			wantToken: "test",
			wantErr:   false,
		},
		{
			name: "error",
			args: args{
				ctx:        context.TODO(),
				sessionID:  "error",
				nextToken:  "error",
				minio:      &mockS3Reader{},
				bucketName: "error",
			},
			want:      nil,
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		got, token, err := getGitOpsRunLogs(
			tt.args.ctx,
			tt.args.sessionID,
			tt.args.nextToken,
			tt.args.minio,
			tt.args.bucketName,
			"")
		g.Expect(err != nil).To(Equal(tt.wantErr))
		g.Expect(got).To(Equal(tt.want))
		g.Expect(token).To(Equal(tt.wantToken))
	}
}
