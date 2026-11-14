package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ValidateProtocol checks the protocol-dependent fields of an endpoint.
func (e Endpoint) ValidateProtocol() error {
	if e.Protocol == nil {
		if e.Operation != nil || e.Variables != nil {
			return errors.New("protocol required for operation or variables")
		}
		return nil
	}

	protocol := *e.Protocol
	var method string
	switch protocol {
	case Gql, Grpc:
		method = http.MethodPost
	case Ws, SSE:
		method = http.MethodGet
		if e.Body != nil {
			return errors.New("live endpoint cannot carry a request body")
		}
		if e.Operation != nil {
			return errors.New("operation requires GraphQL or gRPC")
		}
	default:
		return fmt.Errorf("unknown request protocol %q", protocol)
	}
	if e.Method != method {
		return fmt.Errorf("protocol %q requires method %s", protocol, method)
	}
	if e.Variables != nil {
		if protocol != Gql {
			return errors.New("variables require GraphQL")
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal([]byte(*e.Variables), &object); err != nil || object == nil {
			return errors.New("GraphQL variables must be a JSON object")
		}
	}
	return nil
}
