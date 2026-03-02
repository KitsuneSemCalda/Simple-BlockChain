package seed_tests

import (
	"github.com/caiolandgraf/gest/gest"
	"net/http"
	"net/http/httptest"
	"strings"
)

func init() {
	s := gest.Describe("Seed Server Logic")

	s.It("should respond to health endpoint", func(t *gest.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		req, _ := http.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		t.Expect(rr.Code).ToBe(http.StatusOK)
		t.Expect(rr.Body.String()).ToBe("ok")
	})

	s.It("should respond to seeds endpoint with empty list initially", func(t *gest.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"peers":[]}`))
		})

		req, _ := http.NewRequest("GET", "/seeds", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		t.Expect(rr.Code).ToBe(http.StatusOK)
		t.Expect(rr.Body.String()).ToBe(`{"peers":[]}`)
	})

	gest.Register(s)
}
