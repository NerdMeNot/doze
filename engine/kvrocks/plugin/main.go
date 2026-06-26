// Command kvrocks-plugin runs the kvrocks engine as an out-of-process doze module
// over the engine plugin protocol. Point doze at it with
// DOZE_KVROCKS_PLUGIN=/path/to/kvrocks-plugin.
package main

import (
	"encoding/gob"

	dozeplugin "github.com/nerdmenot/doze-sdk/plugin"
	"github.com/nerdmenot/doze/engine/kvrocks"
)

func main() {
	gob.Register(&kvrocks.Config{})
	dozeplugin.Serve(kvrocks.Driver{})
}
