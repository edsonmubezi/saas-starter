package events

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event
type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	AggregateID   string          `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	Payload       json.RawMessage `json:"payload"`
	Timestamp     time.Time       `json:"timestamp"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

// Handler is a function that handles an event
type Handler func(ctx context.Context, event Event) error

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler Handler)
}

// LocalPublisher is an in-memory event publisher (for monolith)
// TODO: Replace with RabbitMQ/Kafka publisher for microservices
type LocalPublisher struct {
	subscribers map[string][]Handler
	mu          sync.RWMutex
}

// NewLocalPublisher creates a new in-memory event publisher
func NewLocalPublisher() *LocalPublisher {
	return &LocalPublisher{
		subscribers: make(map[string][]Handler),
	}
}

// Publish publishes an event to all subscribers
func (p *LocalPublisher) Publish(ctx context.Context, event Event) error {
	// Set event metadata
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	p.mu.RLock()
	handlers, exists := p.subscribers[event.Type]
	p.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		log.Printf("No subscribers for event type: %s", event.Type)
		return nil
	}

	log.Printf("Publishing event: %s (type: %s) to %d subscribers",
		event.ID, event.Type, len(handlers))

	// Call handlers asynchronously
	for _, handler := range handlers {
		go func(h Handler) {
			if err := h(ctx, event); err != nil {
				log.Printf("Error handling event %s: %v", event.ID, err)
			}
		}(handler)
	}

	return nil
}

// Subscribe registers a handler for a specific event type
func (p *LocalPublisher) Subscribe(eventType string, handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.subscribers[eventType] = append(p.subscribers[eventType], handler)
	log.Printf("Registered handler for event type: %s (total handlers: %d)",
		eventType, len(p.subscribers[eventType]))
}

// NewEvent creates a new event
func NewEvent(eventType, aggregateID, aggregateType string, payload interface{}) (Event, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ID:            uuid.New().String(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Payload:       payloadJSON,
		Timestamp:     time.Now(),
		OccurredAt:    time.Now(),
	}, nil
}

// Example usage:
//
// // In main.go or container
// eventBus := events.NewLocalPublisher()
//
// // Subscribe to events
// eventBus.Subscribe("employee.created", func(ctx context.Context, event events.Event) error {
//     var payload struct {
//         EmployeeID int64  `json:"employee_id"`
//         Email      string `json:"email"`
//     }
//     json.Unmarshal(event.Payload, &payload)
//
//     // Create user account
//     return userService.CreateUser(ctx, payload.Email, payload.EmployeeID)
// })
//
// // In employee use case
// func (u *EmployeeImpl) CreateEmployee(ctx context.Context, emp *Employee) error {
//     // Create employee
//     newEmp, err := u.repo.Create(ctx, emp)
//     if err != nil {
//         return err
//     }
//
//     // Publish event
//     event, _ := events.NewEvent("employee.created",
//         strconv.FormatInt(newEmp.ID, 10),
//         "Employee",
//         map[string]interface{}{
//             "employee_id": newEmp.ID,
//             "email":       newEmp.Email,
//         })
//
//     u.eventPublisher.Publish(ctx, event)
//     return nil
// }