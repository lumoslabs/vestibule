package environ

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

var log Logger = NewLogger("debug", os.Stdout)

type jsonLogger struct {
	zl zerolog.Logger
}

// NewLogger returns a zerologger that conforms to the Logger interface
func NewLogger(l string, w io.Writer) *jsonLogger {
	lvl, er := zerolog.ParseLevel(l)
	if er != nil {
		lvl = zerolog.Disabled
	}
	zerolog.SetGlobalLevel(lvl)
	return &jsonLogger{zerolog.New(w).With().Str("pkg", "environ").Logger()}
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
