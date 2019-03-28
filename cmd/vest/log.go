package main

import (
	"io"

	"github.com/rs/zerolog"
)

type zl struct {
	zerolog.Logger
}

func newLogger(level string, writer io.Writer) *zl {
	zlog := zerolog.New(writer).With().Timestamp().Logger().Level(zerolog.Disabled)
	if lvl, er := zerolog.ParseLevel(level); er == nil {
		zlog = zlog.Level(lvl)
	}
	return &zl{zlog}
}

func (l *zl) Info(msg string) {
	l.Logger.Info().Msg(msg)
}

func (l *zl) Infof(fmt string, objs ...interface{}) {
	l.Logger.Info().Msgf(fmt, objs...)
}

func (l *zl) Debug(msg string) {
	l.Print(msg)
}

func (l *zl) Debugf(fmt string, objs ...interface{}) {
	l.Printf(fmt, objs...)
}
