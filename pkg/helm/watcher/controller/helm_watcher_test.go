package controller

import (
	"context"
	"errors"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/helm/helmfakes"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache/cachefakes"
)

var (
	profile1 = &pb.Profile{
		Name:              "test-profiles-1",
		Home:              "home",
		Description:       "description",
		Icon:              "icon",
		KubeVersion:       "1.21",
		AvailableVersions: []string{"0.0.1", "0.0.2"},
	}
	profile2 = &pb.Profile{
		Name:              "test-profiles-2",
		Home:              "home",
		Description:       "description",
		Icon:              "icon",
		KubeVersion:       "1.21",
		AvailableVersions: []string{"0.0.4"},
	}
)

func TestReconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	repo := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HelmRepository",
			APIVersion: "source.toolkit.fluxcd.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Status: sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Path:     "relative/path",
				URL:      "https://github.com",
				Revision: "revision",
				Checksum: "checksum",
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo)

	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1, profile2}, nil)
	fakeRepoManager.GetValuesFileReturnsOnCall(0, []byte("value1"), nil)
	fakeRepoManager.GetValuesFileReturnsOnCall(1, []byte("value2"), nil)
	fakeRepoManager.GetValuesFileReturnsOnCall(2, []byte("value3"), nil)

	reconciler := &HelmWatcherReconciler{
		Client:      fakeClient.Build(),
		Cache:       fakeCache,
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	expectedData := cache.Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values: map[string]map[string][]byte{
			profile1.Name: {
				"0.0.1": []byte("value1"),
				"0.0.2": []byte("value2"),
			},
			profile2.Name: {
				"0.0.4": []byte("value3"),
			},
		},
	}
	_, namespace, name, cacheData := fakeCache.PutArgsForCall(0)
	assert.Equal(t, "test-namespace", namespace)
	assert.Equal(t, "test-name", name)
	assert.Equal(t, expectedData, cacheData)

	_, helmRepo, ref, filename := fakeRepoManager.GetValuesFileArgsForCall(0)
	assert.Equal(t, repo.Status.Artifact, helmRepo.Status.Artifact)
	assert.Equal(t, &helm.ChartReference{
		Chart:   profile1.Name,
		Version: "0.0.1",
	}, ref)
	assert.Equal(t, "values.yaml", filename)

	_, helmRepo, ref, filename = fakeRepoManager.GetValuesFileArgsForCall(1)
	assert.Equal(t, repo.Status.Artifact, helmRepo.Status.Artifact)
	assert.Equal(t, &helm.ChartReference{
		Chart:   profile1.Name,
		Version: "0.0.2",
	}, ref)
	assert.Equal(t, "values.yaml", filename)

	_, helmRepo, ref, filename = fakeRepoManager.GetValuesFileArgsForCall(2)
	assert.Equal(t, repo.Status.Artifact, helmRepo.Status.Artifact)
	assert.Equal(t, &helm.ChartReference{
		Chart:   profile2.Name,
		Version: "0.0.4",
	}, ref)
	assert.Equal(t, "values.yaml", filename)
}

func TestReconcileGetChartFails(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Status: sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Path:     "relative/path",
				URL:      "https://github.com",
				Revision: "revision",
				Checksum: "checksum",
			},
		},
	}

	fakeRepoManager.ListChartsReturns(nil, errors.New("nope"))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo)
	reconciler := &HelmWatcherReconciler{
		Client:      fakeClient.Build(),
		Cache:       fakeCache,
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}
func TestReconcileGetValuesFileFailsItWillContinue(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Status: sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Path:     "relative/path",
				URL:      "https://github.com",
				Revision: "revision",
				Checksum: "checksum",
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo)

	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1, profile2}, nil)
	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

	reconciler := &HelmWatcherReconciler{
		Client:      fakeClient.Build(),
		Cache:       fakeCache,
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	expectedData := cache.Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values:   map[string]map[string][]byte{},
	}
	_, namespace, name, cacheData := fakeCache.PutArgsForCall(0)
	assert.Equal(t, "test-namespace", namespace)
	assert.Equal(t, "test-name", name)
	assert.Equal(t, expectedData, cacheData)
}

func TestReconcileIgnoreReposWithoutArtifact(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo)

	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1, profile2}, nil)
	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

	reconciler := &HelmWatcherReconciler{
		Client:      fakeClient.Build(),
		Cache:       fakeCache,
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})

	assert.NoError(t, err)
	assert.Zero(t, fakeRepoManager.ListChartsCallCount())
	assert.Zero(t, fakeRepoManager.GetValuesFileCallCount())
	assert.Zero(t, fakeCache.PutCallCount())
}

func TestReconcileUpdateReturnsError(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
		Status: sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Path:     "relative/path",
				URL:      "https://github.com",
				Revision: "revision",
				Checksum: "checksum",
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo)

	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1, profile2}, nil)
	fakeRepoManager.GetValuesFileReturns([]byte("value"), nil)
	fakeCache.PutReturns(errors.New("nope"))

	reconciler := &HelmWatcherReconciler{
		Client:      fakeClient.Build(),
		Cache:       fakeCache,
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

type mockClient struct {
	client.Client
	err error
}

func (m *mockClient) Get(ctx context.Context, key client.ObjectKey, object client.Object) error {
	return m.err
}

func TestReconcileKubernetesGetFails(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	reconciler := &HelmWatcherReconciler{
		Client:      &mockClient{err: errors.New("nope")},
		Cache:       fakeCache,
		RepoManager: fakeRepoManager,
	}
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
	assert.Zero(t, fakeRepoManager.ListChartsCallCount())
	assert.Zero(t, fakeRepoManager.GetValuesFileCallCount())
	assert.Zero(t, fakeCache.PutCallCount())
}
