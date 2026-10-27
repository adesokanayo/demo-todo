package docs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSwaggerDoesNotRedirect(t *testing.T) {
	h := Handler()
	for _, path := range []string{"/swagger", "/swagger/", "/swagger/index.html"} {
		res := httptest.NewRecorder()
		h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, res.Code)
		}
		if !strings.Contains(res.Body.String(), "swagger-ui") {
			t.Fatalf("%s body missing swagger ui", path)
		}
	}

	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("spec status = %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), `"swagger": "2.0"`) {
		t.Fatal("spec is not the OpenAPI document")
	}
}
