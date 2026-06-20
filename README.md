# Order Platform Customer Service

Go service responsible for customer profiles, addresses, contact persons, and
tax profiles. It uses PostgreSQL schema `customer` and validates bearer tokens
through the Auth Service.

## Local development

```powershell
copy .env.example .env
go mod download
go run ./cmd/server
```

The service listens on `http://localhost:3002` and exposes APIs under
`/api/v1`. PostgreSQL from Docker is available to local tools at
`127.0.0.1:15432`.

## Commands

```powershell
go fmt ./...
go vet ./...
go test ./...
```
