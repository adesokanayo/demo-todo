package docs

import (
	"embed"
	"net/http"
)

//go:embed swagger.json index.html
var files embed.FS

// Handler serves Swagger UI and the OpenAPI document.
func Handler() http.Handler {
	index := mustRead("index.html")
	spec := mustRead("swagger.json")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger", "/swagger/", "/swagger/index.html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(index)
		case "/swagger/doc.json", "/swagger/swagger.json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(spec)
		default:
			http.NotFound(w, r)
		}
	})
}

func mustRead(name string) []byte {
	body, err := files.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return body
}
