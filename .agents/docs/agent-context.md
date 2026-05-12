# Agent Context — mgdb-core (MongoDB Core Library)

## Overview
Shared Go library providing reusable MongoDB CRUD patterns for all microservices in the platform. Not a standalone service — imported as a dependency by every other API.
- **Module**: `github.com/devinterop/mgdb-core`
- **Type**: Go library (not a standalone HTTP service)
- **Database**: MongoDB driver wrapper
- **Used by**: All 44+ microservices in the platform

---

## Project Structure

```
mgdb-core/
├── main.go                         # May exist for docs/testing purposes
├── LICENSE
├── README.md
├── go.mod
├── go.sum
├── app/                            # Library implementation
│   └── model_name_db/              # Core CRUD controller pattern
│       ├── controllers/            # InsertController, ReadController, UpdateController, DeleteController
│       ├── services/               # Service layer abstractions
│       └── structs/                # Shared data structures
├── packages/
└── utils/
```

---

## Core Patterns Provided

| Pattern | Description |
|---------|-------------|
| `InsertController` | Standard MongoDB insert (single + bulk) |
| `ReadController` | Standard MongoDB find (one, many, paginated) |
| `UpdateController` | Standard MongoDB update (set, replace) |
| `DeleteController` | Standard MongoDB delete (soft + hard) |

---

## How Services Use This Library

```go
import "github.com/devinterop/mgdb-core/app/model_name_db/controllers"

// Example: Insert a document
result, err := controllers.InsertController(ctx, collection, document)

// Example: Find documents
docs, err := controllers.ReadController(ctx, collection, filter, options)
```

---

## Important Rules for Modifying This Library

1. **Backward compatibility is critical** — all 44+ services import this; breaking changes affect all services
2. **Test thoroughly** — any change must be tested across at least 3+ consuming services
3. **Semver** — bump minor version for new features, patch for bug fixes, major only for breaking changes
4. **Do not add service-specific logic** — keep this library generic/reusable

---

## Versioning

- Check `go.mod` for current version
- All consuming services pin a specific version in their `go.mod`
- After any change, update consuming services incrementally

---

## Key Dependencies

| Package | Role |
|---------|------|
| `go.mongodb.org/mongo-driver` | MongoDB driver |
| `github.com/google/uuid` | ID generation |
