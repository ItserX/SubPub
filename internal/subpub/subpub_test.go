package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSubPub(t *testing.T) {
	sp := NewSubPub()

	t.Run("Basic publish subscribe", func(t *testing.T) {
		var received interface{}
		var wg sync.WaitGroup
		wg.Add(1)

		sub, err := sp.Subscribe("test", func(msg interface{}) {
			received = msg
			wg.Done()
		})
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		err = sp.Publish("test", "hello")
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}

		wg.Wait()
		if received != "hello" {
			t.Errorf("Expected 'hello', got %v", received)
		}

		sub.Unsubscribe()
	})

	t.Run("Multiple subscribers", func(t *testing.T) {
		var wg sync.WaitGroup
		count := 3
		wg.Add(count)

		received := make([]interface{}, count)
		subs := make([]Subscription, count)

		for i := 0; i < count; i++ {
			i := i
			sub, err := sp.Subscribe("test-multi", func(msg interface{}) {
				received[i] = msg
				wg.Done()
			})
			if err != nil {
				t.Fatalf("Failed to subscribe: %v", err)
			}
			subs[i] = sub
		}

		err := sp.Publish("test-multi", "broadcast")
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}

		wg.Wait()
		for i, msg := range received {
			if msg != "broadcast" {
				t.Errorf("Subscriber %d: Expected 'broadcast', got %v", i, msg)
			}
		}

		for _, sub := range subs {
			sub.Unsubscribe()
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		received := false
		sub, _ := sp.Subscribe("test-unsub", func(msg interface{}) {
			received = true
		})

		sub.Unsubscribe()

		err := sp.Publish("test-unsub", "hello")
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
		if received {
			t.Error("Should not receive message after unsubscribe")
		}
	})

	t.Run("Close", func(t *testing.T) {
		sp := NewSubPub()
		messageCount := 0
		var wg sync.WaitGroup
		wg.Add(1)

		sp.Subscribe("test-close", func(msg interface{}) {
			messageCount++
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		})

		sp.Publish("test-close", "last message")

		ctx := context.Background()
		err := sp.Close(ctx)
		if err != nil {
			t.Fatalf("Failed to close: %v", err)
		}

		wg.Wait()
		if messageCount != 1 {
			t.Errorf("Expected 1 message, got %d", messageCount)
		}
	})

	t.Run("Close with timeout", func(t *testing.T) {
		sp := NewSubPub()
		sp.Subscribe("test-timeout", func(msg interface{}) {
			time.Sleep(200 * time.Millisecond)
		})

		sp.Publish("test-timeout", "message")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := sp.Close(ctx)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}
