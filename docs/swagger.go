package docs

import (
	"embed"
	"net/http"
)

//go:embed swagger.json index.html
var files embed.FS

// Handler serves Swagger UI and the OpenAPI document.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger", "/swagger/":
			http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
		case "/swagger/index.html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFileFS(w, r, files, "index.html")
		case "/swagger/doc.json", "/swagger/swagger.json":
			w.Header().Set("Content-Type", "application/json")
			http.ServeFileFS(w, r, files, "swagger.json")
		default:
			http.NotFound(w, r)
		}
	})
}
