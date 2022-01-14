package controller

import (
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestGenerationUpdateReconcilerPredicate_Update(t *testing.T) {
	tests := []struct {
		name  string
		event event.UpdateEvent
		want  bool
	}{
		{
			name:  "returns false no new or old object detected",
			event: event.UpdateEvent{},
			want:  false,
		},
		{
			name: "returns false no old",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
				},
			},
			want: false,
		},
		{
			name: "returns false no new",
			event: event.UpdateEvent{
				ObjectOld: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
				},
			},
			want: false,
		},
		{
			name: "returns true if old source's generation is lower than the new source's",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
				},
				ObjectOld: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
				},
			},
			want: true,
		},
		{
			name: "returns false if old source's generation is higher than the new source's",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
				},
				ObjectOld: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
				},
			},
			want: false,
		},
		{
			name: "returns false if old source's generation equals with the new source's",
			event: event.UpdateEvent{
				ObjectNew: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
				},
				ObjectOld: &sourcev1.HelmRepository{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			he := GenerationUpdatePredicate{}
			assert.Equalf(t, tt.want, he.Update(tt.event), "Update(old: %+v, new: %+v)", tt.event.ObjectOld, tt.event.ObjectNew)
		})
	}
}
