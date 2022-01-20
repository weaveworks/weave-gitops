package controller

import (
	"context"
	"errors"
	"testing"
	"time"

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
	reconciler, fakeCache, _, _ := setupReconcileAndFakes(repo1)
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

func TestReconcileDelete(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, fakeCache, _, _ := setupReconcileAndFakes(repo)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	_, namespace, name := fakeCache.DeleteArgsForCall(0)
	assert.Equal(t, "test-namespace", namespace)
	assert.Equal(t, "test-name", name)
}

func TestReconcileDeletingTheCacheFails(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, fakeCache, _, _ := setupReconcileAndFakes(repo)

	fakeCache.DeleteReturns(errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestReconcileGetChartFails(t *testing.T) {
	reconciler, _, fakeRepoManager, _ := setupReconcileAndFakes(repo1)
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
	reconciler, fakeCache, fakeRepoManager, _ := setupReconcileAndFakes(repo1)
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
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-namespace",
		},
	}
	reconciler, fakeCache, fakeRepoManager, _ := setupReconcileAndFakes(repo)

	fakeRepoManager.GetValuesFileReturns(nil, errors.New("this will be skipped"))

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
	reconciler, fakeCache, _, _ := setupReconcileAndFakes(repo1)
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
	reconciler, fakeCache, _, fakeEventRecorder := setupReconcileAndFakes(repo1)
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.0"}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)

	_, meta, reason, message, _ := fakeEventRecorder.EventInfofArgsForCall(0)

	assert.Equal(t, map[string]string{"revision": "revision"}, meta)
	assert.Equal(t, "info", reason)
	assert.Equal(t, "New version available for profile test-profiles-1 with version 0.0.1", message)
}

func TestDoNotNotifyForLesserOrEqualVersion(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, fakeEventRecorder := setupReconcileAndFakes(repo1)
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.1"}, nil)
	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
	assert.Zero(t, fakeEventRecorder.EventInfofCallCount())
}

func TestNotifyForGreaterVersionListAvailableVersionsReturnsErrorIsSkipped(t *testing.T) {
	reconciler, fakeCache, _, _ := setupReconcileAndFakes()
	fakeCache.ListAvailableVersionsForProfileReturns(nil, errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
}

func TestNotifyForGreaterVersionListAvailableVersionsReturnsHigherVersion(t *testing.T) {
	reconciler, fakeCache, fakeRepoManager, fakeEventRecorder := setupReconcileAndFakes()
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.1"}, nil)
	fakeRepoManager.ListChartsReturns([]*pb.Profile{profile1}, nil)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
	assert.Zero(t, fakeEventRecorder.EventInfofCallCount())
}

func TestNotifyForGreaterVersionEventSenderFailureIsIgnored(t *testing.T) {
	reconciler, fakeCache, _, fakeEventRecorder := setupReconcileAndFakes()
	fakeCache.ListAvailableVersionsForProfileReturns([]string{"0.0.0"}, nil)
	fakeEventRecorder.EventInfofReturns(errors.New("nope"))

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.NoError(t, err)
}

type mockClient struct {
	client.Client
	getErr    error
	updateErr error
	patchErr  error
	obj       *sourcev1.HelmRepository
}

func (m *mockClient) Get(ctx context.Context, key client.ObjectKey, object client.Object) error {
	if m.obj != nil {
		if v, ok := object.(*sourcev1.HelmRepository); ok {
			*v = *m.obj
		}
	}

	return m.getErr
}

func (m *mockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return m.updateErr
}

func (m *mockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return m.patchErr
}

func TestReconcileKubernetesGetFails(t *testing.T) {
	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	reconciler := &HelmWatcherReconciler{
		Client:      &mockClient{getErr: errors.New("nope")},
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

func TestReconcileUpdateFailsDuringDelete(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, _, _, _ := setupReconcileAndFakes()
	reconciler.Client = &mockClient{
		obj:       repo,
		updateErr: errors.New("nope"),
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func TestReconcilePatchFails(t *testing.T) {
	newTime := metav1.NewTime(time.Now())
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-name",
			Namespace:         "test-namespace",
			DeletionTimestamp: &newTime,
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
	reconciler, _, _, _ := setupReconcileAndFakes()
	reconciler.Client = &mockClient{
		obj:      repo,
		patchErr: errors.New("nope"),
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	})
	assert.EqualError(t, err, "nope")
}

func setupReconcileAndFakes(objects ...client.Object) (*HelmWatcherReconciler, *cachefakes.FakeCache, *helmfakes.FakeHelmRepoManager, *controllerfakes.FakeEventRecorder) {
	scheme := runtime.NewScheme()
	utilruntime.Must(sourcev1.AddToScheme(scheme))

	fakeCache := &cachefakes.FakeCache{}
	fakeRepoManager := &helmfakes.FakeHelmRepoManager{}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...)
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
