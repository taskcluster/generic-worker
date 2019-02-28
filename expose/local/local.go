// Local exposer implementation.  Local exposers are useful for local testing,
// and simply exposes the URL as given, or (for ExposeTCPPort) proxies a
// websocket on localhost to the port.

package local

import (
	"fmt"
	"net"

	"net/url"

	"github.com/taskcluster/generic-worker/expose"
	exposeutil "github.com/taskcluster/generic-worker/expose/util"
)

func New(publicIP net.IP) (expose.Exposer, error) {
	return &exposer{publicIP}, nil
}

type exposer struct {
	publicIP net.IP
}

// httpExposure exposes an HTTP server
type httpExposure struct {
	exposer    *exposer
	targetPort uint16
	listener   net.Listener
	proxy      exposeutil.ExposeProxy
}

func (exposer *exposer) ExposeHTTP(targetPort uint16) (expose.Exposure, error) {
	exposure := &httpExposure{exposer: exposer, targetPort: targetPort}
	err := exposure.start()
	if err != nil {
		return nil, err
	}
	return exposure, nil
}

func (exposure *httpExposure) start() error {
	// allocate a port dynamically by specifying :0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	exposure.listener = listener

	proxy, err := exposeutil.ProxyHTTP(listener, exposure.targetPort)
	if err != nil {
		listener.Close()
		return err
	}

	exposure.proxy = proxy
	return nil
}

func (exposure *httpExposure) Close() error {
	if exposure.proxy != nil {
		return exposure.proxy.Close()
	}
	return nil
}

func (exposure *httpExposure) GetURL() *url.URL {
	_, portStr, _ := net.SplitHostPort(exposure.listener.Addr().String())

	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", exposure.exposer.publicIP, portStr),
	}
}

// portExposure exposes a port, via a websocket server
type portExposure struct {
	exposer    *exposer
	targetPort uint16
	listener   net.Listener
	proxy      exposeutil.ExposeProxy
}

func (exposer *exposer) ExposeTCPPort(targetPort uint16) (expose.Exposure, error) {
	exposure := &portExposure{exposer: exposer, targetPort: targetPort}
	err := exposure.start()
	if err != nil {
		return nil, err
	}
	return exposure, nil
}

func (exposure *portExposure) start() error {
	// allocate a port dynamically by specifying :0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	exposure.listener = listener

	proxy, err := exposeutil.ProxyTCPPort(listener, exposure.targetPort)
	if err != nil {
		listener.Close()
		return err
	}

	exposure.proxy = proxy
	return nil
}

func (exposure *portExposure) Close() error {
	if exposure.proxy != nil {
		return exposure.proxy.Close()
	}
	return nil
}

func (exposure *portExposure) GetURL() *url.URL {
	_, portStr, _ := net.SplitHostPort(exposure.listener.Addr().String())

	return &url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%s", exposure.exposer.publicIP, portStr),
	}
}
