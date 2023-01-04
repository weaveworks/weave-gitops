package logger_test

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	l "github.com/weaveworks/weave-gitops/core/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewFromConfigCreatesDevLogger(t *testing.T) {
	g := NewGomegaWithT(t)

	level, err := zapcore.ParseLevel("debug")
	g.Expect(err).NotTo(HaveOccurred())

	dev := true
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(dev),
		l.WithOutAndErrPaths("stdout", "stderr"),
	)

	r, w := redirectStdout(t)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	logline := "wow such log"
	log.Info(logline)

	logs := getLogs(t, r, w)
	bits := strings.Split(string(logs), "\t")
	g.Expect(bits).To(ContainElements(
		MatchRegexp(rfc3339Regexp),
		"gitops",
		logline+"\n",
	))
}

func TestNewFromConfigCreatesProdLogger(t *testing.T) {
	g := NewGomegaWithT(t)

	level, err := zapcore.ParseLevel("debug")
	g.Expect(err).NotTo(HaveOccurred())

	dev := false
	cfg := l.BuildConfig(
		l.WithLogLevel(level),
		l.WithMode(dev),
		l.WithOutAndErrPaths("stdout", "stderr"),
	)

	r, w := redirectStdout(t)

	log, err := l.NewFromConfig(cfg)
	g.Expect(err).NotTo(HaveOccurred())

	logline := "wow such log"
	log.Info(logline)

	var logs map[string]string
	_ = json.Unmarshal(getLogs(t, r, w), &logs)

	g.Expect(logs).To(HaveKeyWithValue("ts", MatchRegexp(epochRegexp)))
	g.Expect(logs).To(HaveKeyWithValue("logger", "gitops"))
	g.Expect(logs).To(HaveKeyWithValue("msg", logline))
}

func TestBuildConfigWithOptions(t *testing.T) {
	g := NewGomegaWithT(t)

	type assert func(zap.Config)

	testCases := []struct {
		name   string
		opt    l.LogOption
		expect assert
	}{
		{
			name: "WithDisableStackTrace should set to true",
			opt:  l.WithDisableStackTrace(),
			expect: func(c zap.Config) {
				g.Expect(c.DisableStacktrace).To(BeTrue())
			},
		},
		{
			name: "WithMode dev=true should configure dev opts",
			opt:  l.WithMode(true),
			expect: func(c zap.Config) {
				g.Expect(c.Development).To(BeTrue())
				g.Expect(c.Encoding).To(Equal("console"))
			},
		},
		{
			name: "WithMode dev=false should configure prod opts",
			opt:  l.WithMode(false),
			expect: func(c zap.Config) {
				g.Expect(c.Development).To(BeFalse())
				g.Expect(c.Encoding).To(Equal("json"))
			},
		},
		{
			name: "WithOutAndErrPaths should append paths to output arrays",
			opt:  l.WithOutAndErrPaths("foo", "bar"),
			expect: func(c zap.Config) {
				g.Expect(c.OutputPaths).To(ContainElement("foo"))
				g.Expect(c.ErrorOutputPaths).To(ContainElement("bar"))
			},
		},
		{
			name: "WithEncoding should set the encoding format",
			opt:  l.WithEncoding("magical"),
			expect: func(c zap.Config) {
				g.Expect(c.Encoding).To(Equal("magical"))
			},
		},
		{
			name: "WithDevelopment should configure dev opts",
			opt:  l.WithDevelopment(),
			expect: func(c zap.Config) {
				g.Expect(c.Development).To(BeTrue())
			},
		},
		{
			name: "WithProduction should configure prod opts",
			opt:  l.WithProduction(),
			expect: func(c zap.Config) {
				g.Expect(c.Development).To(BeFalse())
				g.Expect(c.Sampling.Initial).To(Equal(100))
				g.Expect(c.Sampling.Thereafter).To(Equal(100))
			},
		},
		{
			name: "WithSampling should configure sampling options",
			opt:  l.WithSampling(5, 10),
			expect: func(c zap.Config) {
				g.Expect(c.Sampling.Initial).To(Equal(5))
				g.Expect(c.Sampling.Thereafter).To(Equal(10))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RegisterTestingT(t)

			cfg := l.BuildConfig(tc.opt)
			tc.expect(cfg)
		})
	}
}

var (
	rfc3339Regexp = `^((\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:?\d{2})$`
	epochRegexp   = `^((?:(\d+.\d+))?)$`
)

func redirectStdout(t *testing.T) (*os.File, *os.File) {
	g := NewGomegaWithT(t)
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	g.Expect(err).NotTo(HaveOccurred())

	os.Stdout = w

	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	return r, w
}

func getLogs(t *testing.T, r, w *os.File) []byte {
	g := NewGomegaWithT(t)
	t.Helper()

	w.Close()

	out, err := io.ReadAll(r)
	g.Expect(err).NotTo(HaveOccurred())

	return out
}
