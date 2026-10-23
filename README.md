# Schemo Contracts

Shared protocol definitions, generated bindings, and publication schemas for the **Schemo** desktop client and server.

This repository provides a single source of truth for RPC contracts, publication formats, and cross-language compatibility.

## Repository Structure

| Directory      | Description                                                  |
|----------------|--------------------------------------------------------------|
| `proto/`       | Protocol Buffers definitions and wire compatibility baseline |
| `proto/common/`| Messages and enums shared by several services                |
| `go/`          | Generated Go bindings for Protobuf, gRPC, and Connect        |
| `typescript/`  | Generated ECMAScript modules and TypeScript declarations     |
| `schema/`      | Versioned publication JSON Schemas and test fixtures         |
| `snapshot/v1/` | Generated Go and TypeScript snapshot types                   |
| `secret/`      | Secret paths in sync entity data, shared by client and server |

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

Generated files should remain synchronized with their source definitions. Consumers do not need to run these commands.

## Compatibility

Version 0.5.0 deliberately breaks compatibility with 0.4.x: history records and runs carry the exchange as a JSON document with secret paths, like sync entities, and `common/v1/http.proto` is removed. Upgrade client and server together and reset stored history.

Version 0.4.0 deliberately breaks compatibility with 0.3.x: sync entities carry their content as JSON data with secret paths instead of typed payloads. Upgrade client and server together and reset persisted sync payloads.

Version 0.2.0 deliberately breaks compatibility with 0.1.x: RPC paths, field numbers and enum representations changed. Upgrade client and server together and migrate or explicitly reset persisted protobuf payloads. The publication JSON snapshot remains v1.

Contracts are shared across independently released clients and server. Changes must preserve compatibility with existing consumers.

The following are part of the compatibility surface:

- Protobuf field numbers and wire types.
- Protobuf package and service names.
- JSON field names and publication schema formats.
- Versioned RPC and snapshot contracts.

Avoid reusing removed Protobuf field numbers or introducing incompatible changes within an existing contract version.

Breaking changes should be introduced through explicitly versioned contracts.

## Security Boundaries

This repository defines data structures and communication contracts only. It does not implement application-level authorization or security policies.

The server is responsible for:

- Permission checks and access control.
- Subscription validation and enforcement.
- Credential verification.
- Validation and enforcement of security-sensitive operations.

Clients must not be treated as trusted authorities for these decisions.

## License

Licensed under the [MIT License](LICENSE).

Generated Google error details retain their [Apache-2.0 license](THIRD_PARTY_NOTICES.md).
