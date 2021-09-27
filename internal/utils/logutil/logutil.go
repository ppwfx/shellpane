package logutil

import (
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelPanic = "panic"
	LevelFatal = "fatal"
)

// NewStackDriverLogger creates a *zap.SugaredLogger that is configured for the Stackdriver backend
func NewStackDriverLogger() (sl *zap.SugaredLogger, err error) {
	c := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Encoding:          "json",
		EncoderConfig:     StackdriverEncoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}

	l, err := c.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return &callerCore{Core: c}
	}))
	if err != nil {
		err = errors.Wrap(err, "failed to build zap logger")

		return
	}

	return l.Sugar(), nil
}

//
//func NewDevelopmentLogger() (sl *zap.SugaredLogger, err error) {
//	c := zap.NewDevelopmentEncoderConfig()
//	c.EncodeLevel = zapcore.CapitalColorLevelEncoder
//	l := zap.New(zapcore.NewCore(
//		zapcore.NewConsoleEncoder(c),
//		zapcore.AddSync(colorable.NewColorableStdout()),
//		zapcore.DebugLevel,
//	), zap.WrapCore(func(c zapcore.Core) zapcore.Core {
//		return &callerCore{Core: c}
//	}))
//
//	return l.Sugar(), nil
//}

// LoggerConfig defines opts for creating a new *zap.SugaredLogger
type LoggerConfig struct {
	Backend      string
	MinLevel     string
	TimeFormat   string
	UseColor     bool
	ReportCaller bool
	UseJSON      bool
}

// NewLogger creates a new *zap.SugaredLogger
func NewLogger(opts LoggerConfig) (sl *zap.SugaredLogger, err error) {
	switch strings.ToLower(opts.Backend) {
	case "stackdriver":
		return NewStackDriverLogger()
	}

	if opts.UseColor && opts.UseJSON {
		return nil, errors.New("logutil.LoggerConfig.UseColor and .UseJSON cant be used together")
	}

	var l zapcore.Level
	switch strings.ToLower(opts.MinLevel) {
	case LevelDebug:
		l = zapcore.DebugLevel
	case LevelInfo:
		l = zapcore.InfoLevel
	case LevelWarn:
		l = zapcore.WarnLevel
	case LevelError:
		l = zapcore.ErrorLevel
	case LevelPanic:
		l = zapcore.PanicLevel
	case LevelFatal:
		l = zapcore.FatalLevel
	default:
		l = zapcore.InfoLevel
	}

	var e string
	switch opts.UseJSON {
	case true:
		e = "json"
	default:
		e = "console"
	}

	ec := zap.NewDevelopmentEncoderConfig()

	if opts.TimeFormat != "" {
		ec.EncodeTime = zapcore.TimeEncoderOfLayout(opts.TimeFormat)
	}

	if opts.UseColor {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	c := &zap.Config{
		Level:             zap.NewAtomicLevelAt(l),
		Encoding:          e,
		EncoderConfig:     ec,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
		DisableCaller:     !opts.ReportCaller,
	}

	logger, err := c.Build()
	if err != nil {
		err = errors.Wrap(err, "failed to build zap logger")

		return
	}

	return logger.Sugar(), nil
}

// StackdriverEncoderConfig defines a zapcore.EncoderConfig that is configured for the Stackdriver backend
var StackdriverEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "eventTime",
	LevelKey:       "severity",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "message",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    stackdriverEncodeLevel,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

var stackdriverLogLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "ALERT",
	zapcore.FatalLevel:  "EMERGENCY",
}

func stackdriverEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(stackdriverLogLevelSeverity[lv])
}

type callerCore struct {
	zapcore.Core
}

func (c *callerCore) With(fields []zapcore.Field) zapcore.Core {
	return &callerCore{
		Core: c.Core.With(fields),
	}
}

func (c *callerCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}

	return ce
}

func (c *callerCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	var functionName string
	fn := runtime.FuncForPC(entry.Caller.PC)
	if fn != nil {
		functionName = strings.TrimSuffix(strings.TrimRight(fn.Name(), "0123456789"), ".func")
	}

	fields = append(fields, zap.Object("context.reportLocation", &logReportLocation{
		filePath:     entry.Caller.File,
		lineNumber:   entry.Caller.Line,
		functionName: functionName,
	}))

	return c.Core.Write(entry, fields)
}

type logReportLocation struct {
	filePath     string
	lineNumber   int
	functionName string
}

func (l *logReportLocation) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("filePath", l.filePath)
	e.AddInt("lineNumber", l.lineNumber)
	e.AddString("functionName", l.functionName)
	return nil
}
