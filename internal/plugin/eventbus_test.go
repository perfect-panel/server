package plugin

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestEventBusSubscribePublish(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "test.event", Handler: "on_test"})
	payload, _ := structpb.NewStruct(map[string]interface{}{"key": "val"})
	bus.Publish("test.event", payload)

	// Publish is async (go handler), wait briefly
	time.Sleep(50 * time.Millisecond)

	if n := received.Load(); n != 1 {
		t.Errorf("received = %d, want 1", n)
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "e", Handler: "h1"})
	bus.Subscribe(EventSubscription{PluginName: "p2", Event: "e", Handler: "h2"})
	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "e", Handler: "h3"})

	bus.Publish("e", nil)
	time.Sleep(50 * time.Millisecond)

	if n := received.Load(); n != 3 {
		t.Errorf("received = %d, want 3", n)
	}
}

func TestEventBusFiltersByEvent(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "order.paid", Handler: "h1"})
	bus.Subscribe(EventSubscription{PluginName: "p2", Event: "user.created", Handler: "h2"})

	bus.Publish("order.paid", nil)
	time.Sleep(50 * time.Millisecond)

	if n := received.Load(); n != 1 {
		t.Errorf("received = %d, want 1", n)
	}
}

func TestEventBusSubscribeDedup(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	sub := EventSubscription{PluginName: "p1", Event: "order.paid", Handler: "on_order_paid"}
	bus.Subscribe(sub)
	bus.Subscribe(sub)

	bus.Publish("order.paid", nil)
	time.Sleep(50 * time.Millisecond)

	if n := received.Load(); n != 1 {
		t.Errorf("received = %d, want 1", n)
	}
}

func TestEventBusWildcardPublishDoesNotDoubleDispatch(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "*", Handler: "on_any"})
	bus.Publish("*", nil)
	time.Sleep(50 * time.Millisecond)

	if n := received.Load(); n != 1 {
		t.Errorf("received = %d, want 1", n)
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "e", Handler: "h1"})
	bus.Subscribe(EventSubscription{PluginName: "p2", Event: "e", Handler: "h2"})
	bus.Unsubscribe("p1")

	bus.Publish("e", nil)
	time.Sleep(50 * time.Millisecond)

	if n := received.Load(); n != 1 {
		t.Errorf("received = %d, want 1 (only p2 should remain)", n)
	}
}

func TestEventBusNilHandler(t *testing.T) {
	bus := NewEventBus(nil)
	bus.Subscribe(EventSubscription{PluginName: "p1", Event: "e", Handler: "h1"})
	// Should not panic
	bus.Publish("e", nil)
}

func TestEventBusConcurrent(t *testing.T) {
	var received atomic.Int32
	bus := NewEventBus(func(sub EventSubscription, event string, payload *structpb.Struct) {
		received.Add(1)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bus.Subscribe(EventSubscription{PluginName: "p" + string(rune('a'+idx%26)), Event: "e", Handler: "h"})
		}(i)
	}
	wg.Wait()

	bus.Publish("e", nil)
	time.Sleep(100 * time.Millisecond)
}
