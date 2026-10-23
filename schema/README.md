# Snapshot v1

`snapshot.v1.json` is the source for snapshot data types in server Go, desktop Go and browser
TypeScript. Format/version constants come from its metadata. RPC services remain independently
defined in proto; generating the document model does not make it a service response.

Run `go tool -modfile=tools/go.mod task generate` from the repository root after editing the schema.
`task check` verifies generated types and fixtures. Generation uses pinned quicktype-core and gofmt.
Desktop and server import `github.com/schemo-app/proto/snapshot/v1`; the browser imports types from
`@schemo-app/proto/snapshot/v1`. Consumers do not keep generated copies.

## Writer model and reader compatibility

The schema describes the complete current v1 writer model. Go emits all fields; nullable arrays are
intentional because a nil slice marshals as null. Request body may also be null. Time is RFC3339 JSON:
Go uses time.Time and TypeScript uses string, without runtime Date conversion.

The reader accepts earlier v1 documents which omitted later fields such as prose, changes or the
environment list. Standard Go JSON decoding supplies their zero values, and the view adapter turns
nullable lists into empty arrays. Reader resource budgets, format/version checks and gzip limits remain
authoritative server policy. Wrong primitive shapes are rejected by the generated Go decoder.
Unknown fields retain the prior reader tolerance. The complete writer schema is not applied as a
strict HTTP input validator: doing so would reject already-readable earlier documents. Ajv checks
writer fixtures in tooling; its additionalProperties rule is not a claim about HTTP enforcement.

The schema has no platform secret/token fields. Desktop publication still uses an explicit allowlist
and excludes secret variables. Types alone cannot detect a secret deliberately placed in public text.

Each changelog release carries its actual publication revision. Change items list affected fields,
including parameter location/name and response status/name. `initial:<count>` identifies the first
publication; `restored:<revision>` identifies a restored version. These labels describe public data
only and do not include raw request credentials or response bodies.

## Compatibility evidence

`fixtures/v1.json` anchors the complete writer representation. Server and desktop test the same
expected digest; desktop additionally tests its legacy endpoint fingerprint and imports encoded bytes.
Go struct JSON field order is preserved using quicktype's quicktypePropertyOrder from schema declaration
order. This matters because legacy desktop fingerprints hash endpoint JSON directly. Reordering fields
would create false changelog entries even though the content digest sorts JSON keys.

Do not reorder properties or change field meaning without evaluating compatibility. The golden tests
also cover timestamp representation, null lists, JSON names and zero defaults. Existing negative tests
continue to cover later versions, gzip expansion/resource limits and malformed shapes. Frontend tests
cover early v1 defaults and nested nullable arrays. Changes to an already-shipped version require an
explicit migration/version strategy; updating golden values merely to make a test pass is insufficient.
