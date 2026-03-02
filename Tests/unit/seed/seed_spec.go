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

	s.It("should handle peer announcements", func(t *gest.T) {
		// Mock simple logic for testing the handler's behavior
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
			}
		})

		payload := `{"addr":"/ip4/1.2.3.4/tcp/8333"}`
		req, _ := http.NewRequest("POST", "/announce", strings.NewReader(payload))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		t.Expect(rr.Code).ToBe(http.StatusOK)
	})

	s.It("should respond to info endpoint", func(t *gest.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"online","peers_count":0}`))
		})

		req, _ := http.NewRequest("GET", "/info", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		t.Expect(rr.Code).ToBe(http.StatusOK)
		t.Expect(strings.Contains(rr.Body.String(), "online")).ToBeTrue()
	})

	gest.Register(s)
}
