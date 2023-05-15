package zapstackdriver

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func init() {
	if err := zap.RegisterEncoder("stackdriver", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return &Encoder{
			Encoder: zapcore.NewJSONEncoder(cfg),
		}, nil
	}); err != nil {
		panic(err)
	}
}

func NewStackdriverEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:    "eventTime",
		LevelKey:   "severity",
		NameKey:    "logger",
		CallerKey:  "caller",
		MessageKey: "message",
		//StacktraceKey: "trace",
		LineEnding: zapcore.DefaultLineEnding,
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			case zapcore.DebugLevel:
				enc.AppendString("DEBUG")
			case zapcore.InfoLevel:
				enc.AppendString("INFO")
			case zapcore.WarnLevel:
				enc.AppendString("WARNING")
			case zapcore.ErrorLevel:
				enc.AppendString("ERROR")
			case zapcore.DPanicLevel:
				enc.AppendString("CRITICAL")
			case zapcore.PanicLevel:
				enc.AppendString("ALERT")
			case zapcore.FatalLevel:
				enc.AppendString("EMERGENCY")
			}
		},
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

type Encoder struct {
	zapcore.Encoder
}

func (e *Encoder) Clone() zapcore.Encoder {
	return &Encoder{
		Encoder: e.Encoder.Clone(),
	}
}

func (e *Encoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	if ent.Caller.Defined {
		for _, f := range fields {
			if f.Key == "error" && f.Type == zapcore.ErrorType {
				ent.Message = ent.Message + "\ndue to error: " + f.Interface.(error).Error()
			}
		}
	}

	if ent.Caller.Defined {
		fields = append(fields, zap.Object("logging.googleapis.com/sourceLocation", reportLocation{
			File: ent.Caller.File,
			Line: ent.Caller.Line,
		}))
		ent.Caller.Defined = false
	}

	if ent.Stack != "" {
		ent.Message = ent.Message + "\n" + ent.Stack
		ent.Stack = ""
	}

	return e.Encoder.EncodeEntry(ent, fields)
}

// NewStackdriverConfig ...
func NewStackdriverConfig() zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "stackdriver",
		EncoderConfig:    NewStackdriverEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
}

type ServiceContext struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

// MarshalLogObject implements zapcore ObjectMarshaler.
func (sc ServiceContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if sc.Service == "" {
		return errors.New("service name is mandatory")
	}
	enc.AddString("service", sc.Service)
	enc.AddString("version", sc.Version)

	return nil
}

// LINK: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
type HTTPRequest struct {
	Method    string
	URL       string
	UserAgent string
	Referrer  string
	Status    int
	RemoteIP  string
	Latency   time.Duration
}

// MarshalLogObject implements zapcore ObjectMarshaler.
func (hr HTTPRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("requestMethod", hr.Method)
	enc.AddString("requestUrl", hr.URL)
	enc.AddString("userAgent", hr.UserAgent)
	enc.AddString("referrer", hr.Referrer)
	enc.AddString("remoteIp", hr.RemoteIP)
	enc.AddInt("status", hr.Status)
	enc.AddString("latency", fmt.Sprintf("%gs", hr.Latency.Seconds()))

	return nil
}

type reportLocation struct {
	File     string
	Line     int
	Function string
}

// MarshalLogObject implements zapcore ObjectMarshaler.
func (rl reportLocation) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file", rl.File)
	enc.AddInt("line", rl.Line)
	if rl.Function != "" {
		enc.AddString("function", rl.Function)
	}

	return nil
}
