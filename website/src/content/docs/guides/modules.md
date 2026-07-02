---
title: "Modules, for users"
description: How engines are provided, selected, locked, and upgraded — and every command you need.
---

Every engine except `process` is a **module**: a signed plugin doze fetches
from the registry the first time your config names its type. This page is the
user's side of that machinery — [authors go here](/modules/overview/).

## The two versions (only one is yours)

- **Engine version** — what you declare: `version = 18`. The actual Postgres.
- **Module version** — the plugin's own release, selected automatically: the
  newest release compatible with your doze and every engine version you
  declared, then pinned in `doze.lock`.

You never write a module version. You meet them in `doze modules` output and
in errors that name their own fix.

## The lifecycle, command by command

```sh
doze modules search              # discover: what does the registry publish?
doze modules docs postgres      # the full config reference, in your terminal
doze up                          # first use fetches, verifies, pins — done
doze modules list                # what this project runs, and from where
doze modules info doze/postgres # provenance: releases, protocol, signatures
doze modules upgrade --check    # anything newer that's compatible? (CI: exit 1)
doze modules upgrade             # move the pins; commit the updated doze.lock
```

Pins **never move on their own** — a moving registry can't drift a locked
project, and warm caches resolve fully offline.

## When doze asks you to upgrade

Two situations produce it, both with the exact command in the error:

1. **A new engine major.** You set `version = 19` but the pinned module
   supports 14–18:
   ```
   postgres 19 needs a newer doze/postgres module: pinned 0.2.1 supports
   14, 15, 16, 17, 18 — run 'doze modules upgrade postgres'
   ```
2. **A new config argument.** You used something added in a later module
   release; the decode error names the module and, when a compatible upgrade
   exists, says so.

Some arguments are **version-gated by the engine**, not the module — using a
Postgres-18-only setting with `version = 16` fails at `doze lint` naming the
argument and the required major. The docs mark these (*engine ≥ 18*).

## The `modules {}` block (rarely needed)

```hcl
modules {
  mirror = "file:///path/to/registry"   # air-gapped / dev registry

  cache {
    source  = "acme/valkey"             # a third-party publisher's module
    version = "0.2.0"                   # hold back an exact MODULE release
  }
}
```

Defaults are right for almost everyone: type `postgres` → source
`doze/postgres` → public registry. The `version` knob exists for bisecting a
module regression; the lock is the normal pin. Full field reference:
[configuration → modules](/reference/configuration/#modules).

## Development overrides

- `DOZE_<TYPE>_PLUGIN=/path/to/plugin` — run a local plugin binary, skipping
  the registry entirely (the module-author loop).
- `DOZE_MODULES_MIRROR=…` — point every fetch at another registry base.
- `DOZE_MODULES=off` — no fetching at all (offline, `process`-only).

## Trust, in one paragraph

The registry index that *selects* your module is ed25519-signed; every archive
checksum is signed; the publisher key pins on first use into `doze.lock`.
Tampered, unsigned, or key-rotated ⇒ hard error. The full story:
[the trust model](/why/trust/).
