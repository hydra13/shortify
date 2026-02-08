package auditsaver

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
)

type AuditSaver struct {
	file string
	log  zerolog.Logger
}

func New(file string, log zerolog.Logger) *AuditSaver {
	return &AuditSaver{
		file: file,
		log:  log,
	}
}

func (as *AuditSaver) Handle(event models.Event) {
	bytes, err := json.Marshal(event)
	if err != nil {
		as.log.Error().
			Err(err).
			Interface("event", event).
			Msg("[AuditSaver] can't marshal audit event")
		return
	}

	err = os.WriteFile(as.file, bytes, 0o644)
	if err != nil {
		as.log.Error().
			Err(err).
			Interface("event", event).
			Str("file", as.file).
			Msg("[AuditSaver] can't write event to audit file")
		return
	}

	as.log.Debug().
		Interface("event", event).
		Str("file", as.file).
		Msg("[AuditSaver] successfully save audit event into file")
}
