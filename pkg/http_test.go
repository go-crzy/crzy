package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_NewHTTPListener(t *testing.T) {
	r := &runContainer{
		Log: &mockLogger{},
	}
	v, err := r.newHTTPListener(":8080")
	if err != nil {
		t.Error("should succeed")
	}
	if v == nil {
		t.Error("should not be nil")
		t.FailNow()
	}
	if v.log == nil || !v.log.Enabled() {
		t.Error("log should be enabled")
	}
	if v.errc == nil {
		t.Error("cron should not be empty")
	}
}

func Test_LoggingMiddleware(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
	server := httptest.NewServer(LoggingMiddleware(newCrzyLogger("demo", false), handler))
	client := server.Client()

	request, _ := http.NewRequest("Get", server.URL, nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Status Code should be 200, current: %d",
			response.StatusCode,
		)
	}
}
