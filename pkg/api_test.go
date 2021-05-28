package pkg

import (
	"bytes"
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

func Test_listVersionDetails_succeed(t *testing.T) {
	handler := &verHandler{
		state: &stateManager{state: &mockState{}},
	}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/versions/xxx", nil)
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
	if string(body) != `{"runners": {"deploy": {} }}` {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}

func Test_listVersionDetails_fails(t *testing.T) {
	handler := &verHandler{
		state: &stateManager{state: &mockState{}},
	}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/versions/fail", nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Status Code should be 200, current: %d",
			response.StatusCode,
		)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if string(body) != `{"message":"not found"}` {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}

func Test_action_succeed(t *testing.T) {
	handler := &actionHandler{}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodPost, server.URL+"/v0/actions", bytes.NewBufferString(`{"command":"start"}`))
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
	if string(body) != `{"message":"started"}` {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}

func Test_action_fails_payload(t *testing.T) {
	payloads := []string{`wrong data`, `{"action": "unknown"}`}
	handler := &actionHandler{}
	server := httptest.NewServer(handler)
	client := server.Client()

	for _, v := range payloads {
		request, _ := http.NewRequest(http.MethodPost, server.URL+"/v0/actions", bytes.NewBufferString(v))
		response, err := client.Do(request)
		if err != nil {
			t.Errorf("Should not return %v", err)
		}
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf(
				"Status Code should be 400, current: %d",
				response.StatusCode,
			)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Should not return %v", err)
		}
		if string(body) != `{"message":"bad request"}` {
			t.Errorf("unexpexted data: %s", string(body))
		}
	}
}

func Test_logVersion_succeed(t *testing.T) {
	handler := &verHandler{
		state: &stateManager{state: &mockState{}},
	}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/versions/xxx/log", nil)
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
	if string(body) != "line1\nline2" {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}

func Test_logVersion_fails(t *testing.T) {
	handler := &verHandler{
		state: &stateManager{state: &mockState{}},
	}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/versions/fail/log", nil)
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
	if string(body) != `{"message":"not found"}` {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}

func Test_unknownroute_fails(t *testing.T) {
	handler := &verHandler{
		state: &stateManager{state: &mockState{}},
	}
	server := httptest.NewServer(handler)
	client := server.Client()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/v0/versions/xxx/unknown", nil)
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
	if string(body) != `{"message":"error"}` {
		t.Errorf("Expecting version, get: %s", string(body))
	}
}
