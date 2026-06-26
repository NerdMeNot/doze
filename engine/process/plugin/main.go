// Command process-plugin runs the generic process engine as an out-of-process
// doze module over the engine plugin protocol. Point doze at it with
// DOZE_PROCESS_PLUGIN=/path/to/process-plugin.
package main

import (
	"encoding/gob"

	dozeplugin "github.com/doze-dev/doze-sdk/plugin"
	"github.com/doze-dev/doze/engine/process"
)

func main() {
	gob.Register(&process.Config{})
	dozeplugin.Serve(process.Driver{})
}
