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
	"github.com/weaveworks/weave-gitops/pkg/helm/helmfakes"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache/cachefakes"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/controller/controllerfakes"
)

var (
	profile1 = &pb.Profile{
		Name:              "test-profiles-1",
		Home:              "home",
		Description:       "description",
		Icon:              "icon",
		KubeVersion:       "1.21",
		AvailableVersions: []string{"0.0.1"},
	}
	profile2 = &pb.Profile{
		Name:              "test-profiles-2",
		Home:              "home",
		Description:       "description",
		Icon:              "icon",
		KubeVersion:       "1.21",
		AvailableVersions: []string{"0.0.4"},
	}
	repo1 = &sourcev1.HelmRepository{
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
)

func TestReconcile(t *testing.T) {
	reconciler, fakeCache, _, _ := setupReconcileAndFakes()
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
				"0.0.1": []byte("value"),
			},
			profile2.Name: {
				"0.0.4": []byte("value"),
			},
		},
	}
	_, namespace, name, cacheData := fakeCache.PutArgsForCall(0)
	assert.Equal(t, "test-namespace", namespace)
	assert.Equal(t, "test-name", name)
	assert.Equal(t, expectedData, cacheData)
}

func TestReconcileGetChartFails(t *testing.T) {
	reconciler, _, fakeRepoManager, _ := setupReconcileAndFakes()
	fakeRepoManager.ListChartsReturns(nil, errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}
func TestReconcileGetValuesFileFailsItWillContinue(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, _ := setupReconcileAndFakes()
	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

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

	reconciler, fakeCache, fakeRepoManager, _ := setupReconcileAndFakes()

	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
	}
	reconciler.Client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo).Build()
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
	reconciler, fakeCache, _, _ := setupReconcileAndFakes()
	fakeCache.PutReturns(errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestNotifyForGreaterVersion(t *testing.T) {
	reconciler, fakeCache, _, fakeEventRecorder := setupReconcileAndFakes()
	fakeCache.GetAvailableVersionsForProfileReturns([]string{"0.0.0"}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	_, meta, severity, reason, message, _ := fakeEventRecorder.EventfArgsForCall(0)

	assert.Equal(t, map[string]string{"revision": "revision"}, meta)
	assert.Equal(t, "info", severity)
	assert.Equal(t, "info", reason)
	assert.Equal(t, "New version available for profile test-profiles-1 with version 0.0.1", message)
}

func TestNotifyForGreaterVersionGetAvailableVersionsReturnsErrorIsSkipped(t *testing.T) {
	reconciler, fakeCache, _, _ := setupReconcileAndFakes()
	fakeCache.GetAvailableVersionsForProfileReturns(nil, errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
}

func TestNotifyForGreaterVersionGetAvailableVersionsReturnsHigherVersion(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, fakeEventRecorder := setupReconcileAndFakes()
	fakeCache.GetAvailableVersionsForProfileReturns([]string{"0.0.1"}, nil)
	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
	assert.Zero(t, fakeEventRecorder.EventfCallCount())
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

func setupReconcileAndFakes() (*HelmWatcherReconciler, *cachefakes.FakeCache, *helmfakes.FakeHelmRepoManager, *controllerfakes.FakeEventRecorder) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repo1)
	fakeEventRecorder := &controllerfakes.FakeEventRecorder{}

	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1, profile2}, nil)
	fakeRepoManager.GetValuesFileReturns([]byte("value"), nil)

	return &HelmWatcherReconciler{
		Client:                fakeClient.Build(),
		Cache:                 fakeCache,
		RepoManager:           fakeRepoManager,
		ExternalEventRecorder: fakeEventRecorder,
	}, fakeCache, fakeRepoManager, fakeEventRecorder
}
