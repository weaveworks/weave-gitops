package server

import (
	"fmt"
	"net/http"
	"strings"
)

// WithRoutePrefix wraps the provided mux with a new mux that handles the
// provided route prefix. This is useful when the application is served from a
// subpath, e.g. /weave-gitops. The route prefix is stripped from the request
// before being passed to the provided mux.
//
// We also redirect the user to the route prefix if they hit the root of the
// server, e.g. if they port-forward.
func WithRoutePrefix(mux *http.ServeMux, routePrefix string) *http.ServeMux {
	// ensure that the route prefix begins with a slash
	if !strings.HasPrefix(routePrefix, "/") {
		routePrefix = "/" + routePrefix
	}
	// ensure route prefix doesn't have a trailing slash
	routePrefix = strings.TrimSuffix(routePrefix, "/")

	routePrefixMux := http.NewServeMux()
	routePrefixMux.Handle(routePrefix+"/", http.StripPrefix(routePrefix, mux))
	// Redirect to the route prefix if the user hits the root of the server
	// e.g. if they port-forward
	routePrefixMux.Handle("/", http.RedirectHandler(routePrefix+"/", http.StatusFound))

	return routePrefixMux
}

// GetBaseHref formats the baseHref for the application given the configured route prefix.
// Its used to set the <base href> tag in the index.html file.
// This in turn will form the base URL for all the assets (js, css, etc) that
// are loaded by the browser.
//
// The UI is built with Parcel and has been configured to generate relative script/link tags
// during the build (without a leading or trailing slash) e.g. <script src="main.js">.
// The base href we generate here must complete the URL for these assets and so
// should have a leading and trailing slash, e.g. <base href="/weave-gitops/"> or <base href="/">
func GetBaseHref(routePrefix string) string {
	baseHref := routePrefix
	// ensure baseHref begins and ends with a slash
	if !strings.HasPrefix(baseHref, "/") {
		baseHref = "/" + baseHref
	}
	if !strings.HasSuffix(baseHref, "/") {
		baseHref += "/"
	}
	return baseHref
}

// InjectHTMLBaseTag injects a <base href> tag into the provided html.
// This is used to set the base URL for all the assets (js, css, etc) that
// are loaded by the browser.
func InjectHTMLBaseTag(html []byte, baseHref string) []byte {
	return []byte(strings.Replace(string(html), "<head>", fmt.Sprintf("<head><base href=%q>", baseHref), 1))
}
