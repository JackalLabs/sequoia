package logger

import (
	"github.com/rs/zerolog"
)

type SequoiaLogger struct {
	logger *zerolog.Logger
}

func NewSequoiaLogger(logger *zerolog.Logger) *SequoiaLogger {
	return &SequoiaLogger{logger: logger}
}

func (s *SequoiaLogger) Errorf(format string, i ...any) {
	if len(i) == 0 {
		s.logger.Error().Msg(format)
		return
	}
	s.logger.Error().Msgf(format, i...)
}

func (s *SequoiaLogger) Warningf(format string, i ...any) {
	if len(i) == 0 {
		s.logger.Warn().Msg(format)
		return
	}
	s.logger.Warn().Msgf(format, i...)
}

func (s *SequoiaLogger) Infof(format string, i ...any) {
	if len(i) == 0 {
		s.logger.Info().Msg(format)
		return
	}
	s.logger.Info().Msgf(format, i...)
}

func (s *SequoiaLogger) Debugf(format string, i ...any) {
	if len(i) == 0 {
		s.logger.Debug().Msg(format)
		return
	}
	s.logger.Debug().Msgf(format, i...)
}
