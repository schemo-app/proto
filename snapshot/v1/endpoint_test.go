package snapshot

import "testing"

func TestEndpointProtocol(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		endpoint Endpoint
		valid    bool
	}{
		{name: "older HTTP endpoint", endpoint: Endpoint{Method: "GET"}, valid: true},
		{name: "GraphQL with variables", endpoint: Endpoint{Protocol: new(Gql), Method: "POST", Variables: new(`{"id":7}`)}, valid: true},
		{name: "gRPC with operation", endpoint: Endpoint{Protocol: new(Grpc), Method: "POST", Operation: new("api.Service/Get")}, valid: true},
		{name: "SSE", endpoint: Endpoint{Protocol: new(SSE), Method: "GET"}, valid: true},
		{name: "WebSocket", endpoint: Endpoint{Protocol: new(Ws), Method: "GET"}, valid: true},
		{name: "operation without protocol", endpoint: Endpoint{Method: "GET", Operation: new("Lookup")}},
		{name: "unknown protocol", endpoint: Endpoint{Protocol: new(Protocol("unknown")), Method: "GET"}},
		{name: "incorrect method", endpoint: Endpoint{Protocol: new(Gql), Method: "GET"}},
		{name: "stream with body", endpoint: Endpoint{Protocol: new(SSE), Method: "GET", Body: &Body{}}},
		{name: "stream with operation", endpoint: Endpoint{Protocol: new(Ws), Method: "GET", Operation: new("Listen")}},
		{name: "gRPC with GraphQL variables", endpoint: Endpoint{Protocol: new(Grpc), Method: "POST", Variables: new(`{}`)}},
		{name: "variables array", endpoint: Endpoint{Protocol: new(Gql), Method: "POST", Variables: new(`[]`)}},
		{name: "null variables", endpoint: Endpoint{Protocol: new(Gql), Method: "POST", Variables: new(`null`)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.endpoint.ValidateProtocol(); (err == nil) != test.valid {
				t.Errorf("ValidateProtocol() = %v, want validity %t", err, test.valid)
			}
		})
	}
}
