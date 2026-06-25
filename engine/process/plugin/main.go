// Command process-plugin runs the generic process engine as an out-of-process
// doze module over the engine plugin protocol. Point doze at it with
// DOZE_PROCESS_PLUGIN=/path/to/process-plugin.
package main

import (
	"encoding/gob"

	"github.com/nerdmenot/doze/engine/process"
	dozeplugin "github.com/nerdmenot/doze/internal/plugin"
)

func main() {
	gob.Register(&process.Config{})
	dozeplugin.Serve(process.Driver{})
}
