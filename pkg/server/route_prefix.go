package server

import (
	"fmt"
	"net/http"
	"strings"
)

func WithRoutePrefix(mux *http.ServeMux, routePrefix string) *http.ServeMux {
	// ensure that the route prefix begins with a slash
	if !strings.HasPrefix(routePrefix, "/") {
		routePrefix = "/" + routePrefix
	}
	// ensure route prefix doesn't have a trainling slash
	routePrefix = strings.TrimSuffix(routePrefix, "/")

	routePrefixMux := http.NewServeMux()
	routePrefixMux.Handle(routePrefix+"/", http.StripPrefix(routePrefix, mux))
	// Redirect to the route prefix if the user hits the root of the server
	// e.g. if they port-forward
	routePrefixMux.Handle("/", http.RedirectHandler(routePrefix+"/", http.StatusFound))

	return routePrefixMux
}

func InjectHTMLBaseTag(html []byte, routePrefix string) []byte {
	baseHref := routePrefix
	// ensure baseHref begins and ends with a slash
	if !strings.HasPrefix(baseHref, "/") {
		baseHref = "/" + baseHref
	}
	if !strings.HasSuffix(baseHref, "/") {
		baseHref += "/"
	}
	return []byte(strings.Replace(string(html), "<head>", fmt.Sprintf("<head><base href=%q>", baseHref), 1))
}
