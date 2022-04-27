package logger

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultLogLevel = "info"
	appName         = "gitops"
)

// Levels can be used with log.V() to temporarily reset the configured Level
// Levels are in order of decreasing verbosity and increasing importance
const (
	// LogLevelDebug = -1
	LogLevelDebug int = int(zap.DebugLevel)
	// LogLevelInfo = 0
	LogLevelInfo int = int(zap.InfoLevel)
	// LogLevelWarn = 1
	LogLevelWarn int = int(zap.WarnLevel)
	// LogLevelError = 2
	LogLevelError int = int(zap.ErrorLevel)
)

// New returns a new Logger instance
func New(level string, devMode bool) (logr.Logger, error) {
	config, err := newLoggerConfig(level, devMode)
	if err != nil {
		return logr.Logger{}, err
	}

	return NewFromConfig(config)
}

// NewFromConfig creates a new logr Logger from a zap Config
func NewFromConfig(cfg zap.Config) (logr.Logger, error) {
	log, err := cfg.Build()
	if err != nil {
		return logr.Logger{}, err
	}

	return zapr.NewLogger(log).WithName(appName), nil
}

func newLoggerConfig(level string, devMode bool) (zap.Config, error) {
	l, err := zapcore.ParseLevel(level)
	if err != nil {
		return zap.Config{}, err
	}

	return BuildConfig(
		WithLogLevel(l),
		WithDisableStackTrace(),
		WithMode(devMode),
		WithOutAndErrPaths("stdout", "stderr"),
	), nil
}

// New builds a new Logger Config
func BuildConfig(opts ...LogOption) zap.Config {
	c := zap.Config{}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// LogOption is an option which can be applied to the Logger Config
type LogOption func(*zap.Config)

// WithLogLevel will set the log level on the Logger Config
func WithLogLevel(l zapcore.Level) LogOption {
	return func(c *zap.Config) {
		c.Level = zap.NewAtomicLevelAt(zapcore.Level(l))
	}
}

// WithDisableStackTrace disables the stack traces printed out with Error and
// Warn levels
func WithDisableStackTrace() LogOption {
	return func(c *zap.Config) {
		c.DisableStacktrace = true
	}
}

// WithMode sets up some default options for Dev and Prod envs
func WithMode(devMode bool) LogOption {
	return func(c *zap.Config) {
		WithProduction()(c)
		WithEncoding("json")(c)

		if devMode {
			WithDevelopment()(c)
			WithEncoding("console")(c)
		}
	}
}

// WithOutAndErrPaths appends locations for stdout and stderr. This can be
// called as often as needed
func WithOutAndErrPaths(o, e string) LogOption {
	return func(c *zap.Config) {
		c.OutputPaths = append(c.OutputPaths, o)
		c.ErrorOutputPaths = append(c.OutputPaths, e)
	}
}

// WithEncoding sets the encoding for the logger (eg "console" or "json"
func WithEncoding(e string) LogOption {
	return func(c *zap.Config) {
		c.Encoding = e
	}
}

// WithDevelopment sets some Development mode defaults
func WithDevelopment() LogOption {
	return func(c *zap.Config) {
		c.Development = true
		c.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}
}

// WithProduction sets some Production mode defaults
func WithProduction() LogOption {
	return func(c *zap.Config) {
		c.Development = false
		c.EncoderConfig = zap.NewProductionEncoderConfig()
		WithSampling(100, 100)(c)
	}
}

// WithSampling sets sampling config
func WithSampling(i, t int) LogOption {
	return func(c *zap.Config) {
		c.Sampling = &zap.SamplingConfig{
			Initial:    i,
			Thereafter: t,
		}
	}
}

// WithHumanTimeCode can be used to set more human readable time on Production
// logs
func WithHumanTimeCode() LogOption {
	return func(c *zap.Config) {
		c.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	}
}
