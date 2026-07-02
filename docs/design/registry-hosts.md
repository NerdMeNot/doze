# Design: registry hosts in module sources

**Status:** accepted, not yet implemented. This document fixes the design so
nothing we ship in the meantime forecloses it.

## The problem

A module source is `<namespace>/<name>` (`doze/postgres`, `acme/redis`), which
*reads* multi-publisher — per-namespace ed25519 keys, TOFU pinning, "third-party
modules render identically." But the CLI has exactly **one registry base**
(default `https://doze.nerdmenot.in/registry`, overridable only globally via
`DOZE_MODULES_MIRROR` / `modules { mirror = … }`). Every namespace must
therefore live inside one registry deployment:

- A third party can't host their own registry *and* coexist with the official
  one in the same project — the mirror override is all-or-nothing.
- Publishing under `acme/*` today means PRing `keys.json` and every signed
  index update into the doze-registry repo, with the doze maintainers as merge
  gatekeepers. That's a curated catalog, not an open ecosystem.

Terraform hit the same wall and solved it in the source address:
`registry.terraform.io/hashicorp/aws` with the host optional and defaulted.

## The design

### Source grammar

```
source := [host "/"] namespace "/" name
host   := a DNS name containing at least one "."   (this disambiguates: namespaces never contain dots)
```

- `postgres` block type → `doze/postgres` (unchanged default).
- `modules { cache { source = "acme/redis" } }` — the **default host**
  (unchanged behavior; still served by the official registry if that's where
  `acme` publishes).
- `modules { cache { source = "registry.acme.dev/acme/redis" } }` — the module
  index, keys.json, and catalog entry come from `https://registry.acme.dev`
  (registry layout unchanged: `<host>/<ns>/keys.json`,
  `<host>/<ns>/<name>/index.yaml`).

The dot rule makes parsing unambiguous and keeps every existing source string
valid — this is a pure extension.

### Resolution

- Per-host `binaries.Manager` scoping, exactly as per-namespace scoping works
  today: cache under `~/.doze/modules/<host>/<ns>` (default host keeps its
  current `~/.doze/modules/<ns>` path so nothing re-downloads).
- `DOZE_MODULES_MIRROR` and `modules { mirror = … }` keep their meaning:
  they override *the default host only*. Explicit hosts in sources are not
  redirected — a project that names `registry.acme.dev` means it. (An
  air-gapped-everything switch can come later if real demand shows up;
  per-host mirror maps are complexity we refuse until then.)

### Trust

- TOFU key pinning becomes host-qualified: `doze.lock`'s `keys:` map keys
  change from `<ns>` to `<host>/<ns>`, with bare `<ns>` read as the default
  host (so existing locks stay valid; they rewrite on next save).
- The `modules:` pin layer already keys by full source string — it inherits
  host-qualification for free when sources grow hosts.
- Signature scheme, index schema, artifact verification: unchanged. A registry
  is still just static files; any host serving the layout works.

### Catalog / discovery

`doze modules search` searches the default host's catalog (unchanged). Explicit
hosts are for *use*, not discovery — a third party advertises their source
string themselves. Cross-host search federation is explicitly out of scope.

## What this buys

- A third party runs `keygen`, publishes static files anywhere (Pages, S3, a
  GitHub Pages branch), and their users write one `modules {}` line. No PR to
  doze-registry, no gatekeeping, same security posture.
- The official registry stays the curated default with zero behavior change
  for everyone who never types a host.

## What we deliberately defer

- Implementation (it's ~a day when we want it: source parsing, host-scoped
  managers, lock key qualification, docs).
- Per-host mirror overrides, cross-host search, host allowlists/denylists.
- Any change to the registry file layout or signing — none is needed.

## Guardrails until implemented

- Nothing may assume a source has exactly one `/` (parse with `SplitN`-style
  logic, reject only what's actually invalid).
- Nothing may key persistent state by bare namespace where a host could later
  qualify it, except `doze.lock keys:` which has the compat path above.
- Error messages should keep saying "the registry" generically, not baking in
  the official hostname as the only possibility.
