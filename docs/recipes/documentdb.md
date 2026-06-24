# Recipes — DocumentDB (MongoDB wire)

DocumentDB is doze's way to run a **MongoDB-compatible** document store locally
without MongoDB itself or its SSPL license (see **[The
engines](../guide/engines.md)**). It speaks the MongoDB wire protocol, so MongoDB
drivers, `mongosh`, and GUIs like Compass all work.

It's a single, **self-contained** engine: doze quietly runs a private PostgreSQL 18
carrying Microsoft's DocumentDB extension behind a FerretDB gateway, and exposes
only the Mongo wire protocol. There's no version to pick and no backend to wire up
— the Postgres and the gateway are an implementation detail.

## A Mongo-compatible store

```hcl
documentdb "docs" {}
```

That's the whole declaration. `MONGODB_URI` is injected for the `docs` instance.

```sh
doze run -- sh -c 'mongosh "$MONGODB_URI" --eval "db.runCommand({ping:1})"'
```

> **First boot is slow.** doze builds the cluster and runs `CREATE EXTENSION` the
> first time `docs` boots (a few minutes). After that it's a normal lazy engine —
> sub-second cold boots. Warm it ahead of time with `doze start docs`.

## Use it like Mongo

```sh
eval "$(doze env)"
mongosh "$MONGODB_URI" --eval '
  db.users.insertOne({ name: "Ada", roles: ["admin"] });
  printjson(db.users.find().toArray());
'
```

Point a driver at the same URI:

```js
// Node — the standard mongodb driver, unchanged
new MongoClient(process.env.MONGODB_URI)
```

```python
# Python — pymongo
pymongo.MongoClient(os.environ["MONGODB_URI"])
```

## Connecting a GUI

Find the endpoint and connect MongoDB Compass (or any Mongo client) to it:

```sh
doze status
#   NAME   ENGINE       STATE   …   ENDPOINT
#   docs   documentdb   idle        127.0.0.1:6441
```

Or open `mongosh` directly — `doze shell` picks the right client for the engine:

```sh
doze shell docs
```

## Notes

- DocumentDB targets broad MongoDB compatibility, not 100% — check the
  [FerretDB docs](https://docs.ferretdb.io/) if a specific command matters.
- It is **versionless**: the Postgres + extension + gateway are a curated bundle
  doze pins as a unit, so a `documentdb` block takes no `version`.
