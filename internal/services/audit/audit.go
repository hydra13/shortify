// Модуль работы с событиями аудита.
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
	subscribers map[Subscriber]struct{}
}

func New() *AuditService {
	return &AuditService{
		mu:          &sync.Mutex{},
		subscribers: make(map[Subscriber]struct{}),
	}
}

func (as *AuditService) publishEvent(event models.Event) {
	as.mu.Lock()
	defer as.mu.Unlock()
	for sub := range as.subscribers {
		go sub.Handle(event)
	}
}

func (as *AuditService) Subscribe(subscriber Subscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.subscribers[subscriber] = struct{}{}
}

func (as *AuditService) Unsubscribe(subscriber Subscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()

	delete(as.subscribers, subscriber)
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
