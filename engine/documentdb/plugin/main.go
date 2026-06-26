// Command documentdb-plugin runs the documentdb engine as an out-of-process doze
// module — a composite (private Postgres + DocumentDB extension, fronted by
// FerretDB) served over the engine plugin protocol. Point doze at it with
// DOZE_DOCUMENTDB_PLUGIN=/path/to/documentdb-plugin.
package main

import (
	"encoding/gob"

	dozeplugin "github.com/nerdmenot/doze-sdk/plugin"
	"github.com/nerdmenot/doze/engine/documentdb"
)

func main() {
	gob.Register(&documentdb.Config{})
	dozeplugin.Serve(documentdb.Driver{})
}
