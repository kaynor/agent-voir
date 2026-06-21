module github.com/agentvoir/agentvoir/apps/gateway

go 1.25

require (
	github.com/agentvoir/agentvoir/packages/auth-go v0.0.0
	github.com/alicebob/miniredis/v2 v2.33.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.5.1
)

replace github.com/agentvoir/agentvoir/packages/auth-go => ../../packages/auth-go

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/go-oidc/v3 v3.11.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
)
