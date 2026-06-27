package process

import (
	"fmt"
	"net"

	"github.com/zclconf/go-cty/cty"

	"github.com/doze-dev/doze-sdk/engine"
)

// Attributes implements engine.Attributer: expose the app's listen address as an
// http URL so other instances can reference process.<name>.url. host/port come
// from the generic baseline (derived from the declared port); this adds the URL.
func (Driver) Attributes(_ engine.Instance, ep engine.Endpoint) map[string]cty.Value {
	host, port, err := net.SplitHostPort(ep.TCPAddr)
	if err != nil {
		return nil
	}
	return map[string]cty.Value{
		"url": cty.StringVal(fmt.Sprintf("http://%s:%s", host, port)),
	}
}
