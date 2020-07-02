package utils

import "github.com/rs/zerolog/log"

// HandleMessageError log err in the right format
func HandleMessageError(subject string, err error) {
	log.Error().
		Str("subject", subject).
		Err(err).
		Msg("")
}
