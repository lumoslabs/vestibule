package environ

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

var log = NewLogger("info", os.Stdout)

type jsonLogger struct {
	zl zerolog.Logger
}

// NewLogger returns a zerologger that conforms to the Logger interface
func NewLogger(l string, w io.Writer) Logger {
	lvl, er := zerolog.ParseLevel(l)
	if er != nil {
		lvl = zerolog.Disabled
	}

	return &jsonLogger{zerolog.New(w).With().Timestamp().Logger().Level(lvl)}
}

func (l *jsonLogger) Info(msg string) {
	l.zl.Info().Msg(msg)
}

func (l *jsonLogger) Infof(fmt string, objs ...interface{}) {
	l.zl.Info().Msgf(fmt, objs...)
}

func (l *jsonLogger) Debug(msg string) {
	l.zl.Print(msg)
}

func (l *jsonLogger) Debugf(fmt string, objs ...interface{}) {
	l.zl.Printf(fmt, objs...)
}

// SetLogger will set the package logger
func SetLogger(logger Logger) {
	log = logger
}

// GetLogger gets the package logger
func GetLogger() Logger {
	return log
}
