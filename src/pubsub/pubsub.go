package pubsub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Subscription represents a single subscription to a topic
type Subscription struct {
	ID      string
	Topic   string
	Channel chan interface{}
}

// PubSub manages publish and subscribe operations
type PubSub struct {
	id            uuid.UUID
	mu            sync.RWMutex
	subscriptions map[string]map[string]*Subscription
}

// NewPubSub creates a new PubSub instance
func NewPubSub() *PubSub {
	return &PubSub{
		id:            uuid.New(),
		subscriptions: make(map[string]map[string]*Subscription),
	}
}

// Subscribe creates a new subscription to a specific topic
func (ps *PubSub) Subscribe(ctx context.Context, topic string) (*Subscription, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Generate a unique subscription ID
	subID := generateUniqueID()

	// Create a new subscription
	subscription := &Subscription{
		ID:      subID,
		Topic:   topic,
		Channel: make(chan interface{}, 100), // Buffered channel
	}

	// Ensure topic exists in subscriptions map
	if ps.subscriptions[topic] == nil {
		ps.subscriptions[topic] = make(map[string]*Subscription)
	}

	log.Println("Broker-ID: [", ps.id, "] : new-subscriber:", subscription)
	// Store the subscription
	ps.subscriptions[topic][subID] = subscription

	// Optional: Handle context cancellation
	go func() {
		<-ctx.Done()
		ps.Unsubscribe(topic, subID)
	}()

	return subscription, nil
}

// Publish sends a message to all subscribers of a specific topic
func (ps *PubSub) Publish(topic string, message interface{}) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Check if topic exists
	topicSubscribers, exists := ps.subscriptions[topic]
	if !exists {
		return fmt.Errorf("no subscribers for topic: %s", topic)
	}

	// Broadcast to all subscribers
	for _, sub := range topicSubscribers {
		select {
		case sub.Channel <- message:
			log.Println("Broker-ID: [", ps.id, "] : published-event:", message)
			// Message sent successfully
		default:
			// Channel is full, skip this message
			fmt.Printf("Warning: Channel for subscription %s is full\n", sub.ID)
		}
	}

	return nil
}

// Unsubscribe removes a specific subscription
func (ps *PubSub) Unsubscribe(topic, subID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check if topic exists
	topicSubscribers, exists := ps.subscriptions[topic]
	if !exists {
		return fmt.Errorf("topic not found: %s", topic)
	}

	// Remove the subscription
	if sub, exists := topicSubscribers[subID]; exists {
		close(sub.Channel)
		delete(topicSubscribers, subID)
	}

	// Clean up empty topic maps
	if len(topicSubscribers) == 0 {
		delete(ps.subscriptions, topic)
	}

	log.Println("Broker-ID: [", ps.id, "] : unsubscribe", "topic: ", topic, "| subscriber-id: ", subID)
	return nil
}

// generateUniqueID creates a unique identifier for subscriptions
func generateUniqueID() string {
	// In a real-world scenario, use a more robust unique ID generation method
	// This is a simple placeholder
	return fmt.Sprintf("sub_%d", time.Now().UnixNano())
}

// Example usage and demonstration
// func ExamplePubSubUsage() {
// 	ctx := context.Background()
// 	pubsub := NewPubSub()

// 	log.Println(pubsub)

// 	// Subscribe to a topic
// 	subscription, err := pubsub.Subscribe(ctx, "user_updates")
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Receive messages
// 	go func() {
// 		for {
// 			c := <-subscription.Channel
// 			log.Println(c)
// 			// select {
// 			// case c := <-subscription.Channel:
// 			// 	log.Println(c)
// 			// }
// 		}
// 	}()
// 	// Publish a message
// 	go func() {
// 		userUpdate := map[string]interface{}{
// 			"id":     "user123",
// 			"name":   "John Doe",
// 			"action": "profile_updated",
// 		}
// 		err := pubsub.Publish("user_updates", userUpdate)
// 		if err != nil {
// 			fmt.Println("Publish error:", err)
// 		}
// 	}()
// }
