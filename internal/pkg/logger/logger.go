package logger

import (
	"os"

	"github.com/rs/zerolog"

	"loadbalancer/internal/pkg/config"
)

type Logger interface {
	Infof(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Debugf(msg string, v ...interface{})
}

type ZerologLogger struct {
	logger zerolog.Logger
}

func NewZerologLogger(cfg config.Logger) *ZerologLogger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	//output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &ZerologLogger{logger: logger}
}

func (l *ZerologLogger) Infof(msg string, v ...interface{}) {
	l.logger.Info().Msgf(msg, v...)
}

func (l *ZerologLogger) Errorf(msg string, v ...interface{}) {
	l.logger.Error().Msgf(msg, v...)
}

func (l *ZerologLogger) Debugf(msg string, v ...interface{}) {
	l.logger.Debug().Msgf(msg, v...)
}
