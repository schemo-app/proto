# v1 wire baseline

`v1.binpb` is the descriptor image of the v1 contracts, captured without source info so that comment
edits cannot change it. It includes imports and preserves protobuf service, message, field, enum and
JSON descriptors. `buf.wire.yaml` selects the WIRE_JSON checks for it. Run `task wire` from the
repository root.

This baseline starts at the deliberately incompatible v0.5.0 release, which replaced the typed sync
and history payloads with JSON documents. Passing it does not establish compatibility with earlier releases. Deploy the updated client and server together; migrate or explicitly reset
persisted protobuf payloads before decoding them with the new contracts.

Keep this baseline immutable after v0.5.0 while these contracts are supported. Field numbers may
have gaps. Reserve removed field numbers and names instead of renumbering existing fields.

The baseline was captured from `proto/` with:

```bash
go tool -modfile=../tools/go.mod buf build --exclude-source-info -o compatibility/v1.binpb
```

Release-specific source compatibility must additionally be checked against a published tag using
the regular FILE rules in `buf.yaml`.
