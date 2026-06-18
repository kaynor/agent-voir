package usage

import "context"

// Store persists and queries usage events.
type Store interface {
	Insert(ctx context.Context, event Event) error
	List(ctx context.Context, filter ListFilter) ([]Event, error)
}
