// Package audit - модуль работы с событиями аудита.
package audit

import (
	"sync"
	"time"

	"github.com/hydra13/shortify/internal/models"
)

const (
	EventTypeShorten = "shorten"
	EventTypeFollow  = "follow"
)

type Subscriber interface {
	Handle(event models.Event)
}

// AuditService реализует pub/sub паттерн для публикации событий аудита.
type AuditService struct {
	mu          sync.Locker
	subscribers []Subscriber
}

func New() *AuditService {
	return &AuditService{
		mu:          &sync.Mutex{},
		subscribers: make([]Subscriber, 0),
	}
}

func (as *AuditService) publishEvent(event models.Event) {
	as.mu.Lock()
	defer as.mu.Unlock()
	for _, sub := range as.subscribers {
		go sub.Handle(event)
	}
}

func (as *AuditService) Subscribe(subscriber Subscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.subscribers = append(as.subscribers, subscriber)
}

func (as *AuditService) Unsubscribe(subscriber Subscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()
	for i, s := range as.subscribers {
		if s == subscriber {
			as.subscribers = append(as.subscribers[:i], as.subscribers[i+1:]...)
			break
		}
	}
}

func (as *AuditService) PublishShortenEvent(longURL string, userID string) {
	as.publishEvent(models.Event{
		Timestamp:   time.Now().Unix(),
		Action:      EventTypeShorten,
		UserID:      userID,
		OriginalURL: longURL,
	})
}

func (as *AuditService) PublishFollowEvent(longURL string, userID string) {
	as.publishEvent(models.Event{
		Timestamp:   time.Now().Unix(),
		Action:      EventTypeFollow,
		UserID:      userID,
		OriginalURL: longURL,
	})
}
