package pkg

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_newReverseProxy_with_404(t *testing.T) {
	h := newReverseProxy(&mockUpstream{})
	server := httptest.NewServer(h)
	client := server.Client()

	request, _ := http.NewRequest("Get", server.URL, nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Status Code should be 404, current: %d",
			response.StatusCode,
		)
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		t.Error("body conversion should succed")
	}
	body := strings.Split(string(b), "\n")[0]
	if body != `{"message": "NotFound"}` {
		t.Errorf("message should NotFound, >%s<", body)
	}
}

func Test_defaultUpstream(t *testing.T) {
	u := newUpstream(&defaultState{})
	_, err := u.getDefault()
	if err != errServiceNotFound {
		t.Errorf("should returm errServiceNotFound, returns %v", err)
	}
	u.setDefault("localhost:8090")
	h, err := u.getDefault()
	if err != nil {
		t.Errorf("should succeed, returns %v", err)
	}
	if h != "localhost:8090" {
		t.Errorf("should return localhost:8090, returns %v", h)
	}
}

func Test_mockUpstream(t *testing.T) {
	u := mockUpstream{}
	_, err := u.getDefault()
	if err != errServiceNotFound {
		t.Errorf("should returm errServiceNotFound, returns %v", err)
	}
	u.setDefault("localhost:8090")
}
