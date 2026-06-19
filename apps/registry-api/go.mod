module github.com/agentvoir/agentvoir/apps/registry-api

go 1.25

require (
	github.com/agentvoir/agentvoir/packages/auth-go v0.0.0
	github.com/golang-migrate/migrate/v4 v4.17.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/agentvoir/agentvoir/packages/auth-go => ../../packages/auth-go

require (
	github.com/coreos/go-oidc/v3 v3.11.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/rogpeppe/go-internal v1.15.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)
