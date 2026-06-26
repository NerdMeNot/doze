// Command postgres-plugin runs the postgres engine as an out-of-process doze
// module over the engine plugin protocol. Point doze at it with
// DOZE_POSTGRES_PLUGIN=/path/to/postgres-plugin.
package main

import (
	"encoding/gob"

	dozeplugin "github.com/nerdmenot/doze-sdk/plugin"
	"github.com/nerdmenot/doze/engine/postgres"
)

func main() {
	gob.Register(&postgres.Config{})
	dozeplugin.Serve(postgres.Driver{})
}
