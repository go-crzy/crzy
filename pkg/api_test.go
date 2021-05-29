package pkg

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type sample struct {
	name   string
	method string
	route  string
	input  string
	status int
	output string
}

var data = []sample{
	{
		name:   `get_on_configuration_and_succeed`,
		method: http.MethodGet,
		route:  "/v0/configuration",
		input:  "{}",
		status: http.StatusOK,
		output: `{"message":"bad request"}`,
	},
	{
		name:   `put_on_configuration_and_fail`,
		method: http.MethodPut,
		route:  "/v0/configuration",
		input:  `wrong`,
		status: http.StatusBadRequest,
		output: `{"message":"bad request"}`,
	},
	{
		name:   `put_on_configuration_and_succeed`,
		method: http.MethodPut,
		route:  "/v0/configuration",
		input:  `{"head": "main"}`,
		status: http.StatusOK,
		output: `{"message":"bad request"}`,
	},
	{
		name:   `get_on_unknowversionssubcommand_and_fail`,
		method: http.MethodGet,
		route:  "/v0/versions/xxx/unknown",
		input:  ``,
		status: http.StatusOK,
		output: `{"message":"error"}`,
	},
	{
		name:   `get_on_versions_and_fail`,
		method: http.MethodGet,
		route:  "/v0/versions/fail/log",
		input:  ``,
		status: http.StatusNotFound,
		output: `{"message":"not found"}`,
	},
	{
		name:   `get_on_versions_and_succeed`,
		method: http.MethodGet,
		route:  "/v0/versions/xxx/log",
		input:  ``,
		status: http.StatusOK,
		output: "line1\nline2",
	},
	{
		name:   `post_on_action_and_fails_due_to_payload`,
		method: http.MethodPost,
		route:  "/v0/actions",
		input:  `wrong data`,
		status: http.StatusBadRequest,
		output: `{"message":"bad request"}`,
	},
	{
		name:   `post_on_action_and_fails_due_to_payload`,
		method: http.MethodPost,
		route:  "/v0/actions",
		input:  `{"action": "unknown"}`,
		status: http.StatusBadRequest,
		output: `{"message":"bad request"}`,
	},
	{
		name:   `post_on_action_and_succeed`,
		method: http.MethodPost,
		route:  "/v0/actions",
		input:  `{"command": "start"}`,
		status: http.StatusOK,
		output: `{"message":"started"}`,
	},
	{
		name:   `get_on_one_version_and_fails`,
		method: http.MethodGet,
		route:  "/v0/versions/fail",
		input:  ``,
		status: http.StatusNotFound,
		output: `{"message":"not found"}`,
	},
	{
		name:   `get_on_one_version_and_succeeds`,
		method: http.MethodGet,
		route:  "/v0/versions/xxx",
		input:  ``,
		status: http.StatusOK,
		output: `{"runners": {"deploy": {} }}`,
	},
	{
		name:   `get_on_version_and_succeeds`,
		method: http.MethodGet,
		route:  "/v0/version",
		input:  ``,
		status: http.StatusOK,
		output: `version`,
	},
	{
		name:   `get_on_versions_and_succeeds`,
		method: http.MethodGet,
		route:  "/v0/versions",
		input:  ``,
		status: http.StatusOK,
		output: `{"versions": ["123"]}`,
	},
}

func Test_configuration_success(t *testing.T) {
	mux := newAPI(&stateManager{state: &mockState{}})
	server := httptest.NewServer(mux)
	client := server.Client()

	for _, v := range data {
		t.Logf("testing %s", v.name)
		payload := &bytes.Buffer{}
		if len(v.input) > 0 {
			payload = bytes.NewBuffer([]byte(v.input))
		}
		request, _ := http.NewRequest(v.method, server.URL+v.route, payload)
		response, err := client.Do(request)
		if err != nil {
			t.Errorf("Should not return %v", err)
		}
		if response.StatusCode != v.status {
			t.Errorf(
				"status should be %d, current: %d",
				v.status,
				response.StatusCode,
			)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("body should work %v", err)
		}
		if string(body) != v.output {
			t.Errorf("expect %s, get: %s", v.output, string(body))
		}
	}
}
