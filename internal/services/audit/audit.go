package audit

import (
	"sync"
	"time"
)

const (
	EventTypeShorten = "shorten"
	EventTypeFollow  = "follow"
)

type Event struct {
	Timestamp   int64  `json:"ts"`
	Action      string `json:"action"`
	UserID      string `json:"user_id"`
	OriginalURL string `json:"original_url"`
}

type Subscriber interface {
	Notify(event Event)
}

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

func (s *AuditService) publishEvent(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sub := range s.subscribers {
		go sub.Notify(event)
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
	as.publishEvent(Event{
		Timestamp:   time.Now().Unix(),
		Action:      EventTypeShorten,
		UserID:      userID,
		OriginalURL: longURL,
	})
}

func (as *AuditService) PublishFollowEvent(longURL string, userID string) {
	as.publishEvent(Event{
		Timestamp:   time.Now().Unix(),
		Action:      EventTypeFollow,
		UserID:      userID,
		OriginalURL: longURL,
	})
}
