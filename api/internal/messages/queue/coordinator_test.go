package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCoordinatorEnqueueWhileDraining(t *testing.T) {
	t.Parallel()

	c := &Coordinator{}
	c.setDraining(true)

	_, err := c.Enqueue(context.Background(), uuid.New(), map[string]string{"foo": "bar"}, time.Second)
	if err == nil || err != ErrQueueStopped {
		t.Fatalf("expected ErrQueueStopped, got %v", err)
	}
}
