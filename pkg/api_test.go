package pkg

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiVersion(t *testing.T) {
	handler := &versionHandler{}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/version", nil)
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if string(body) != "version" {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}

func TestApiVersions(t *testing.T) {
	handler := &versionsHandler{
		state: &stateManager{state: &mockState{}},
	}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/versions", nil)
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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if string(body) != `{"versions": ["123"]}` {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}
