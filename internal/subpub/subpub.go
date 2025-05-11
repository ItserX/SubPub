package subpub

import (
	"context"
	"sync"
)

type subscription struct {
	subject  string
	messages chan interface{}
	sp       *subPub
	done     chan struct{}
}

type subPub struct {
	subscribers map[string][]*subscription
	mu          *sync.RWMutex
	wg          *sync.WaitGroup
}

const (
	bufferSize = 1024
)

func (s *subscription) Unsubscribe() {
	s.sp.mu.Lock()
	defer s.sp.mu.Unlock()

	if subs, exists := s.sp.subscribers[s.subject]; exists {
		for i, sub := range subs {
			if sub == s {
				s.sp.subscribers[s.subject] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		if len(s.sp.subscribers[s.subject]) == 0 {
			delete(s.sp.subscribers, s.subject)
		}
	}
	close(s.done)
}

func NewSubPub() SubPub {
	return &subPub{
		subscribers: make(map[string][]*subscription),
		mu:          &sync.RWMutex{},
		wg:          &sync.WaitGroup{},
	}
}

func (sp *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sub := &subscription{
		subject:  subject,
		messages: make(chan interface{}, bufferSize),
		sp:       sp,
		done:     make(chan struct{}),
	}

	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		for {
			select {
			case msg, ok := <-sub.messages:
				if !ok {
					return
				}
				cb(msg)
			case <-sub.done:
				return
			}
		}
	}()

	sp.subscribers[subject] = append(sp.subscribers[subject], sub)
	return sub, nil
}

func (sp *subPub) Publish(subject string, msg interface{}) error {
	sp.mu.RLock()
	subs := sp.subscribers[subject]
	sp.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.messages <- msg:
		default:

			continue
		}
	}
	return nil
}

func (sp *subPub) Close(ctx context.Context) error {
	sp.mu.Lock()
	for _, subs := range sp.subscribers {
		for _, sub := range subs {
			close(sub.messages)
		}
	}
	sp.subscribers = make(map[string][]*subscription)
	sp.mu.Unlock()

	done := make(chan struct{})
	go func() {
		sp.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
