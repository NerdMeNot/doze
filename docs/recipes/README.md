# doze recipes

Practical, copy-pasteable examples of what doze can do. Every recipe is a
**config** (`doze.hcl`) plus the **commands** to use it, with a note on what you
get.

New to doze? Start with **[Getting started](../guide/getting-started.md)** and
**[Core concepts](../guide/concepts.md)** first — these recipes assume you know
the basic loop (declare → `doze run` → boots on connect, reaps when idle). For
what each engine *is* and when to reach for it, see **[The
engines](../guide/engines.md)**.

## How every recipe works

1. **Declare** instances in `doze.hcl` (or split across sibling `*.doze.hcl`).
2. **Use** them — every instance listens on the explicit `port` you declared, so
   its URL is stable and deterministic. Connect a client directly to that endpoint
   (doze boots the instance on the first connection and reaps it when idle):
   ```sh
   doze run -- <your command>     # ensures backends are up, then runs the command
   ```
   Three honest ways to get a connection string:
   - write the stable URL straight into your app config / `.env`
     (e.g. `postgresql://app:app@127.0.0.1:5432/app`, `redis://127.0.0.1:6379`,
     `mongodb://127.0.0.1:27017/`) — connecting cold-boots the instance;
   - declare your app as a **`process`** block — doze injects each dependency's
     `env_var` → URL automatically;
   - read **`.doze/endpoints.yaml`**, the machine manifest the daemon writes.

For a `process` block, the conventional `env_var` per engine is:

| Engine | Conventional var |
|---|---|
| postgres | `DATABASE_URL` |
| valkey / kvrocks | `REDIS_URL` |
| documentdb | `MONGODB_URI` |
| s3 / sqs / sns | `AWS_ENDPOINT_URL_S3` / `_SQS` / `_SNS` (+ dummy `AWS_*` creds) |

doze converges **structure** (databases, roles, schemas, grants, extensions,
buckets, queues, topics) — never data. Your app/migrations own the data.

## Index

- [PostgreSQL](postgres.md) — roles, schemas, grants, extensions, multiple DBs, tuning, versions
- [Valkey & Kvrocks](valkey-kvrocks.md) — Redis-protocol cache and durable KV
- [DocumentDB](documentdb.md) — MongoDB wire, self-contained (Postgres + gateway)
- [S3](s3.md) — local object storage (buckets, multipart, presigned URLs)
- [SQS](sqs.md) — queues, FIFO, DLQ + redrive
- [SNS](sns.md) — topics, SNS→SQS fanout, filter policies, webhooks
- [Workflows](workflows.md) — `run`, reset for clean test DBs, status/dash/logs, CI
- [Config layout](config-layout.md) — splitting config across `*.doze.hcl` files + per-dev overrides
- [Full stacks](stacks.md) — polyglot apps end to end + framework wiring

For where doze stores engines, data, sockets, and logs — and what to commit vs
ignore — see the **[Files & storage guide](../guide/files-and-storage.md)**.
