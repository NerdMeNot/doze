---
title: "How modules work"
description: The driver contract, the plugin protocol, and the author's view of the two version axes.
---

A doze module is a **pure-Go plugin binary** implementing the
[doze-sdk](https://github.com/doze-dev/doze-sdk) driver contract over gRPC
(HashiCorp go-plugin). doze core is an engine-blind host — proxy, supervisor,
config evaluator, lockfile — and *everything engine-specific lives in your
module*. Third-party modules are exactly as capable as official ones; there is
no privileged path. This is the Terraform provider model.

## The contract

Your driver is a **stateless value** implementing `engine.Driver`:

| Method | Job |
|---|---|
| `Type()` | The config block keyword (`httpd "site" { … }`) and registry name. |
| `Resolve(spec, platform, lock, fetch)` | Locate/download the **engine binary** for the declared version; honor the lock pin. |
| `Provision(inst, toolchain)` | Make the data dir bootable (run initdb, lay down files). Idempotent. |
| `Provisioned(dataDir)` | Is it already initialized? |
| `Plan(inst, toolchain)` | Return a declarative **SpawnPlan** — core runs and supervises it, gated on your readiness probe. |
| `BackendSocket(socketDir, port)` | The unix socket core's proxy dials. |
| `ConnString(inst, endpoint)` | The URL surfaced to users (`DATABASE_URL`, …). |

Everything richer is an **opt-in capability interface**, discovered by type
assertion and advertised over the wire — implement only what your engine
needs:

- `ConfigDecoder` — decode your own HCL block (strict schema; the declared
  engine version arrives so you can gate arguments — see
  [Describe](/modules/describe/)).
- `Converger` / `Inventory` / `Pruner` — declared structure (roles, buckets…)
  converged into the running engine, tracked and pruned.
- `Templater` — copy-on-write data-dir templates (postgres's instant clones).
- `ProxyFilter` — own your wire protocol's preamble (TLS, startup, cancel);
  runs *in your process* via file-descriptor handoff.
- `Lifecycle`, `HealthChecker`, `Restartable`, `PortBinder` — supervised
  long-lived processes.
- `Versionless` — engines that ship their own server (no `version =`).
- `Describer` — **mandatory in practice**: docs and the signed engine-support
  list generate from it.

The plugin's `main` is three lines: register your config type with gob, call
`dozeplugin.Serve(Driver{})`, and optionally dispatch a hidden `__serve` mode
if your module *is* its own server (doze's S3/SQS/SNS work this way).

## The two axes, from the author's side

Users declare **engine versions**; your module has its own **release version**.
Your `Describe().Versions` list becomes the signed index's engine-support gate:
declare `{"14"…"18"}` and a user writing `version = 19` gets told to upgrade
(or that no release supports it) *before anything runs*. Shipping support for
Postgres 19 is: extend `Describe().Versions` (plus whatever `Resolve`/converge
work it needs), bump your module version, release. Locked projects are
untouched until their owners run `doze modules upgrade`.

The **plugin protocol version** (an SDK constant) is the third compatibility
rail: a doze only selects releases speaking its protocol, and older doze
versions keep selecting your older releases — which is why published releases
are immutable and never deleted.

## The pieces you'll touch

| | |
|---|---|
| `doze-sdk/engine` | The contract: interfaces + value types. |
| `doze-sdk/plugin` | `Serve()`, the gRPC transport, protocol version. |
| `doze-sdk/modtool` | Build/package/index/meta — [releasing](/modules/releasing/). |
| `doze-sdk/enginetest` | Boot a real backend from your driver in tests. |
| [module-template](https://github.com/doze-dev/module-template) | A complete working module to copy. |

Start building: [your first module](/modules/first-module/).
