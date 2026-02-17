package capture

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsPermanentPublishError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "maximum payload exceeded",
			err:  errors.New("nats: maximum payload exceeded"),
			want: true,
		},
		{
			name: "message size exceeds maximum allowed",
			err:  errors.New("nats publish event abc-123: message size exceeds maximum allowed"),
			want: true,
		},
		{
			name: "marshal event envelope",
			err:  fmt.Errorf("marshal event envelope: json: unsupported type: chan int"),
			want: true,
		},
		{
			name: "wrapped maximum payload exceeded",
			err:  fmt.Errorf("nats publish event evt-123: %w", errors.New("maximum payload exceeded")),
			want: true,
		},
		{
			name: "transient connection closed",
			err:  errors.New("nats: connection closed"),
			want: false,
		},
		{
			name: "transient timeout",
			err:  errors.New("nats: timeout"),
			want: false,
		},
		{
			name: "transient no responders",
			err:  errors.New("nats: no responders"),
			want: false,
		},
		{
			name: "generic network error",
			err:  errors.New("dial tcp: connection refused"),
			want: false,
		},
		{
			name: "context deadline exceeded",
			err:  fmt.Errorf("publish: %w", errors.New("context deadline exceeded")),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPermanentPublishError(tt.err)
			if got != tt.want {
				t.Errorf("isPermanentPublishError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
