package testutil

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/piotrkowalczuk/zapstackdriver"
)

type Opts struct {
	Environment string
	Level       string
}

// Init allocates new logger based on given options.
func Init(opts Opts) (logger *zap.Logger, err error) {
	var (
		cfg     zap.Config
		options []zap.Option
		lvl     zapcore.Level
	)
	switch opts.Environment {
	case "production":
		cfg = zap.NewProductionConfig()
	case "stackdriver":
		cfg = zapstackdriver.NewStackdriverConfig()
	case "development":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()
	}

	if err = lvl.Set(opts.Level); err != nil {
		return nil, err
	}
	cfg.Level.SetLevel(lvl)
	cfg.OutputPaths = []string{}

	logger, err = cfg.Build(options...)
	if err != nil {
		return nil, err
	}
	logger.Info("logger has been initialized", zap.String("environment", opts.Environment))

	return logger, nil
}
