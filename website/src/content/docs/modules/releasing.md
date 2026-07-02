---
title: "Releasing modules"
description: modtool builds the archives, the schema-1 index, and the docs — under rules that keep users' locks resolving forever.
---

Releases are built by `doze-sdk/modtool` — the same library the official
monorepo uses, so your release layout is theirs. The template's
`cmd/release/main.go` is the whole integration:

```go
m := modtool.Module{
    Name:       "httpd",
    Version:    *version,          // the MODULE version — bump every change
    Namespace:  "acme",            // your publisher namespace
    PluginPath: "./plugin",
    Driver:     httpd.Driver{},    // Describe() is mandatory
}
err := modtool.Release(m, "dist", modtool.AllTriples())
```

`go run ./cmd/release --version 0.1.0` produces, under `dist/httpd/`:

```
httpd-plugin-0.1.0-aarch64-apple-darwin.tar.gz      # + both Linux triples
index.yaml    # schema-1: per-release protocol, engine-support, artifacts, channels
meta.yaml     # the docs, generated from Describe()
```

The index is written **unsigned** — signing happens at the registry, where the
key lives ([publishing](/modules/publishing/)).

## The rules (enforced, not suggested)

1. **Bump the module version for every change.** New config argument, bug fix,
   new engine major — every one is a new release. Users' locks pin your
   releases; the version is the changelog.
2. **Published artifacts are immutable.** modtool *skips* building any
   `(version, platform)` already present in the index, and hard-errors if a
   rebuild would produce different bytes. Why so strict: a changed published
   checksum strands every `doze.lock` that pinned the original — and Go
   builds embed VCS state, so "the same code" from a later commit isn't the
   same bytes. Never republish; bump.
3. **Merge cumulatively.** Download your previously published `index.yaml`
   into `dist/<name>/` before building (the template's CI does) so old
   releases stay in the index — an older doze, or a project locked last year,
   keeps resolving. The `stable` channel automatically tracks your highest
   version.
4. **Protocol is stamped for you** from the SDK you compiled against. When the
   plugin protocol ever bumps, older doze versions keep selecting your older
   releases — another reason rule 3 matters.

## CI

The template ships a release workflow: dispatch it with a version → tests →
`cmd/release` for all platforms → upload `dist/<name>/*` to a rolling GitHub
release (with upload retries — the release API throws transient 5xxs). That
release is your **archive host**; it needs no trust, because everything users
run is verified against the signed registry index. Which is the next page.
