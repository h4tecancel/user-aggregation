package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"

	sweetLogger "github.com/h4tecancel/sweet-logger"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(env string) (*slog.Logger, func()) {
	switch strings.ToLower(env) {
	case "prod":
		zl := newZapProd()

		handler := slogzap.
			Option{
			Level:     slog.LevelInfo,
			Logger:    zl,
			AddSource: true,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Value = slog.StringValue(time.Now().Format(time.RFC3339Nano))
				}
				return a
			},
		}.
			NewZapHandler()

		l := slog.New(handler).With(
			"app", "user_aggregation",
			"env", env,
		)
		cleanup := func() { _ = zl.Sync() }
		return l, cleanup

	default:
		l := sweetLogger.New(sweetLogger.Options{
			Level:      slog.LevelDebug,
			AddSource:  true,
			TimeFormat: "2006-01-02 15:04:05",
			Color:      sweetLogger.ColorAuto,
			Writer:     os.Stderr,
		}).With("app", "user_aggregation", "env", env)

		return l, func() {}
	}
}

func newZapProd() *zap.Logger {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encCfg.TimeKey = "ts"

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encCfg,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	z, _ := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	return z.With(
		zap.String("app", "user_aggregation"),
		zap.String("env", "prod"),
	)
}
