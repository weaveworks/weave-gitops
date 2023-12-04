package server

import (
	"io"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/go-logr/logr"
)

// AssetHandler returns a http.Handler that serves static assets from the provided fs.FS.
// It also redirects all non-file requests to index.html.
func AssetHandler(assetHandler, redirectHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Assume anything with a file extension in the name is a static asset.
		extension := filepath.Ext(req.URL.Path)
		// We use the golang http.FileServer for static file requests.
		// This will return a 404 on normal page requests, ie /some-page.
		// Redirect all non-file requests to index.html, where the JS routing will take over.
		if extension == "" {
			redirectHandler.ServeHTTP(w, req)
			return
		}
		assetHandler.ServeHTTP(w, req)
	}
}

// IndexHTMLHandler ensures that index.html always gets served.
// The JS router will take care of actual navigation once the index.html page lands.
func IndexHTMLHandler(fsys fs.FS, log logr.Logger, routePrefix string) http.HandlerFunc {
	baseHref := GetBaseHref(routePrefix)
	log.Info("Creating redirector", "routePrefix", routePrefix, "baseHref", baseHref)

	return func(w http.ResponseWriter, r *http.Request) {
		indexPage, err := fsys.Open("index.html")
		if err != nil {
			log.Error(err, "could not open index.html page")
			http.Error(w, "could not open index.html page", http.StatusInternalServerError)
			return
		}
		defer indexPage.Close()

		bt, err := io.ReadAll(indexPage)
		if err != nil {
			log.Error(err, "could not read index.html")
			http.Error(w, "could not read index.html", http.StatusInternalServerError)
			return
		}

		// inject base tag into index.html
		bt = InjectHTMLBaseTag(bt, baseHref)

		_, err = w.Write(bt)
		if err != nil {
			log.Error(err, "error writing index.html")
			http.Error(w, "error writing index.html", http.StatusInternalServerError)
			return
		}
	}
}
