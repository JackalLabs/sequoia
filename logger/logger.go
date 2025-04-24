package logger

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type SequoiaLogger struct{}

func (s *SequoiaLogger) Errorf(format string, i ...any) {
	if len(i) == 0 {
		log.Error().Msg(format)
		return
	}
	log.Error().Msg(fmt.Sprintf(format, i...))
}

func (s *SequoiaLogger) Warningf(format string, i ...any) {
	if len(i) == 0 {
		log.Warn().Msg(format)
		return
	}
	log.Warn().Msg(fmt.Sprintf(format, i...))
}

func (s *SequoiaLogger) Infof(format string, i ...any) {
	if len(i) == 0 {
		log.Info().Msg(format)
		return
	}
	log.Info().Msg(fmt.Sprintf(format, i...))
}

func (s *SequoiaLogger) Debugf(format string, i ...any) {
	if len(i) == 0 {
		log.Debug().Msg(format)
		return
	}
	log.Debug().Msg(fmt.Sprintf(format, i...))
}
