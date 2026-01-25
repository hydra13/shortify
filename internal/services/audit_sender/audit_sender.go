package auditsender

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/hydra13/shortify/internal/models"
)

type AuditSender struct {
	url string
	log zerolog.Logger
}

func New(url string, log zerolog.Logger) *AuditSender {
	return &AuditSender{
		url: url,
		log: log,
	}
}

func (as *AuditSender) Handle(event models.Event) {
	jsonEvent, err := json.Marshal(event)
	if err != nil {
		as.log.Error().
			Err(err).
			Interface("event", event).
			Msg("[AuditSender] can't marshal audit event")
		return
	}

	response, err := http.Post(as.url, "application/json", bytes.NewBuffer(jsonEvent))
	if err != nil {
		as.log.Error().
			Err(err).
			Interface("event", event).
			Str("url", as.url).
			Msg("[AuditSender] can't send audit event")
	}
	defer response.Body.Close()

	as.log.Debug().
		Interface("event", event).
		Str("url", as.url).
		Int("response_status_code", response.StatusCode).
		Msg("[AuditSender] successfully sent audit event")
}
