package cache

import "context"

type Entry struct {
	Key          string
	Value        []byte
	TTLSeconds   int64
	CacheStatus  string
	AgentID      string
	RequestHash  string
	ResponseHash string
}

type Store interface {
	Get(ctx context.Context, key string) (*Entry, error)
	Set(ctx context.Context, entry Entry) error
	Delete(ctx context.Context, key string) error
}
