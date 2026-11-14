# Schemo Proto

Shared protocol definitions, generated bindings, and publication schemas for the **Schemo** desktop client and server.

This repository provides a single source of truth for RPC contracts, publication formats, and cross-language compatibility.

## Repository Structure

| Directory       | Description                                                   |
|-----------------|---------------------------------------------------------------|
| `proto/`        | Protocol Buffers definitions and wire compatibility baseline  |
| `proto/common/` | Messages and enums shared by several services                 |
| `go/`           | Generated Go bindings for Protobuf, gRPC, and Connect         |
| `typescript/`   | Generated ECMAScript modules and TypeScript declarations      |
| `schema/`       | Versioned publication JSON Schemas and test fixtures          |
| `snapshot/v1/`  | Generated Go and TypeScript snapshot types                    |
| `secret/`       | Secret paths in sync entity data, shared by client and server |

## Usage

Consumers should use released artifacts directly. Code generation is only required when modifying contracts.

### Go

Install the versioned module with `go get github.com/schemo-app/proto@v0.5.0`.

Import generated RPC packages:

```go
import "github.com/schemo-app/proto/go/<service>/v1"
```

Import snapshot types:

```go
import "github.com/schemo-app/proto/snapshot/v1"
```

### TypeScript

Install the ready package from the matching GitHub release:

```bash
npm install https://github.com/schemo-app/proto/releases/download/v0.5.0/schemo-app-proto-0.5.0.tgz
```

Import generated RPC bindings:

```typescript
import { ... } from '@schemo-app/proto/rpc/<service>/v1/<file>_pb.js'
```

Import snapshot types:

```typescript
import type { ... } from '@schemo-app/proto/snapshot/v1'
```

## Development

Install dependencies:

```bash
npm ci
```

Generate bindings and snapshot types:

```bash
go tool -modfile=tools/go.mod task generate
```

Verify generated artifacts and contract consistency:

```bash
go tool -modfile=tools/go.mod task check
```


## License

Licensed under the [MIT License](LICENSE).
