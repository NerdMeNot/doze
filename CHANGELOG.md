# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims to
follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

doze is a lazy, no-Docker local runtime for backing services — `proto` for
databases and AWS services. A generic, engine-agnostic core drives per-engine
drivers behind a small interface; each declared instance boots on first connect
and reaps when idle.

### Added

- **Database engines**: PostgreSQL, Valkey, Kvrocks (Redis protocol), and
  FerretDB (MongoDB wire), each a declarative config block (`postgres "n" {}`,
  `valkey "n" {}`, …).
- **Built-in AWS services**: local **S3, SQS, and SNS** implemented in pure Go
  and shipped inside the binary — no Docker, no JVM, no LocalStack. SQS speaks
  both wire protocols (AWS JSON 1.0 + legacy Query/XML) with FIFO and dead-letter
  redrive; SNS does filter policies, raw delivery, SNS→SQS fanout, and HTTP(S)
  webhooks; S3 embeds gofakes3 (buckets, multipart, presigned URLs). All are
  SDK-verified.
- **Engine-agnostic core**: a `Driver` contract plus optional capability
  interfaces (convergence, protocol filter, copy-on-write templates, config
  decode, backend provider, dependencies, env injection, versionless) discovered
  by type assertion. Adding an engine is a driver package plus one registration.
- **Per-instance proxy**: one listener per instance; a connection lazily boots
  the instance (coalesced via `singleflight`) and splices byte-for-byte. Reaping
  keys on connection count, never query inactivity.
- **`doze run` / `doze env`**: ensure instances up and inject connection strings
  (`DATABASE_URL`, `REDIS_URL`, `MONGODB_URI`, `AWS_ENDPOINT_URL_*` + dummy
  creds, plus `DOZE_<NAME>_URL`); `.doze/endpoints.yaml` manifest.
- **Instance dependencies**: a dependent boots and holds its dependencies
  (FerretDB → Postgres backend, SNS → SQS instance), releasing them on stop.
- **Copy-on-write templates & `doze ephemeral`**: `initdb` once per version,
  clone per instance (CoW); throwaway, isolated databases per test run.
- **Multi-file config**: `doze.hcl` + a merged `doze.d/*.hcl` overlay (or
  `--config <dir>`), with positioned, file/line config diagnostics and
  "did you mean?" hints.
- **Interactive TUI** (`doze dash`): select an instance and boot/reap/restart it
  or tail its logs, with a live-updating table.
- **Resilience**: backend-crash detection (mark reaped → clean reboot), bounded
  daemon shutdown, macOS orphan reclamation, and boot/convergence errors surfaced
  in `doze status`/`doctor`.
- **Binaries**: per-engine append-only release mirror, content-addressed cache, a
  committed `doze.lock` (versions + checksums), `DOZE_<ENGINE>_BINDIR` /
  `DOZE_<ENGINE>_MIRROR` (and `file://` mirrors), and `doze versions`. Built and
  published by the companion `doze-binaries` repo.
- **Postgres specifics**: declarative roles/users, schemas, grants, and
  extensions (contrib + prebuilt + from-source bundles); query cancellation via
  the pgbouncer cancel dance; client-facing TLS termination via a `tls {}` block.
- `~/.doze` home laid out like moonrepo's proto: shared per-engine tool stores
  plus per-project, namespaced state. Overridable via `DOZE_HOME` / `data_dir`.
- Daemon lifecycle (`start`/`stop`/`restart [instance]`/`serve`/`logs`), plus
  `init`, `up`, `down`, `status`/`ls`, `psql`, `doctor`, and styled CLI output.
- Licensed under Apache 2.0.

[Unreleased]: https://github.com/NerdMeNot/doze/commits/main
