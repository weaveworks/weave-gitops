package run_test

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/ui/run"
)

func TestMissingTLSKeyOrCert(t *testing.T) {
	log := logrus.New()
	err := run.ListenAndServe(&http.Server{}, false, "foo", "", log)
	assert.ErrorIs(t, err, cmderrors.ErrNoTLSCertOrKey)

	err = run.ListenAndServe(&http.Server{}, false, "", "bar", log)
	assert.ErrorIs(t, err, cmderrors.ErrNoTLSCertOrKey)
}
