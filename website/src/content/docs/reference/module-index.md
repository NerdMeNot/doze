---
title: "Module index schema"
description: The signed schema-1 index a registry serves per module — fields, signatures, selection.
---

Each module's `index.yaml` (served at `<registry>/<ns>/<name>/index.yaml`) is
the machine contract between publishers and doze clients. Schema 1:

```yaml
schema: 1
module: postgres                # must equal the directory name
namespace: doze                 # the publisher namespace (the signing identity)
releases:
  "0.2.1":
    protocol: 1                 # doze plugin protocol this binary speaks
    engines: ["14", "15", "16", "17", "18"]   # engine MAJORS supported;
                                #   absent/empty = versionless, no gate
    artifacts:
      aarch64-apple-darwin:
        url: https://…/postgres-plugin-0.2.1-aarch64-apple-darwin.tar.gz
        sha256: 64607…          # lowercase hex
        sig: 8AKRDh…            # ed25519 over the hex sha256, base64
      aarch64-unknown-linux-gnu: { … }
      x86_64-unknown-linux-gnu: { … }
channels:
  stable: "0.2.1"               # npm-dist-tag style pointer into releases
signature: 5FLx3w…              # ed25519 over the canonical payload (below)
```

## The index signature

`signature` covers the canonical-JSON serialization of
`{module, namespace, releases, channels}`: object keys sorted at every level,
no insignificant whitespace, empty optional fields omitted, `sha256` values
lowercased. Sign = ed25519 over the lowercase-hex SHA-256 of those bytes.
Reference implementations are byte-compatible in Go
(`doze-sdk/modindex.CanonicalPayload/Sign/Verify`) and JS (the registry's
`scripts/lib.mjs`).

It exists so the *selection metadata* is attestable: without it, a hostile
host could forge engine-support claims, roll `channels.stable` back, or
withhold releases — while every individual artifact still verified.

## Selection (what clients do with it)

Implemented once in `doze-sdk/modindex.Select`:

1. Filter releases to those whose `protocol` equals the client's.
2. Filter to those whose `engines` cover every declared engine major
   (empty list = no gate).
3. Prefer the `stable` channel head if it survived the filters; otherwise the
   highest surviving version (numeric dotted compare) — this is how an older
   doze keeps resolving after the channel moves to a newer protocol.

Failures are specific by design: "every release requires protocol ≥ N —
upgrade doze" vs "no release supports postgres 19; latest supports 14–18".

## Invariants

- **Releases are append-only** and published artifacts immutable — clients'
  locks pin these checksums forever.
- A `(version, triple)` entry's bytes never change; new triples may merge into
  an existing release.
- An index without `schema: 1` is rejected ("re-publish the module").
- `meta.yaml` beside the index is generated docs (prose) — deliberately
  unsigned and never load-bearing; everything enforcing behavior lives in the
  signed index.
