package contracts_test

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/schemo-app/proto/go/platform/v1"
	"github.com/schemo-app/proto/go/platform/v1/platformconnect"
)

const exampleServerName = "Example"

type platformHandler struct{}

func (platformHandler) GetInfo(context.Context, *platform.GetInfoRequest) (*platform.GetInfoResponse, error) {
	return platform.GetInfoResponse_builder{
		Server: platform.ServerInfo_builder{Name: exampleServerName, Version: "dev"}.Build(),
	}.Build(), nil
}

func requireExampleServer(t *testing.T, response *platform.GetInfoResponse) {
	t.Helper()
	if response.GetServer().GetName() != exampleServerName {
		t.Fatalf("unexpected server: %v", response)
	}
}

// TestDesktopGRPCAndPanelConnect verifies both clients against the same Connect handler.
func TestDesktopGRPCAndPanelConnect(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(platformconnect.NewPlatformServiceHandler(platformHandler{}))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	t.Run("desktop grpc", func(t *testing.T) {
		roots := x509.NewCertPool()
		roots.AddCert(server.Certificate())
		connection, err := grpc.NewClient(strings.TrimPrefix(server.URL, "https://"),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(roots, "")))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = connection.Close() })
		response, err := platform.NewPlatformServiceClient(connection).GetInfo(ctx, &platform.GetInfoRequest{})
		if err != nil {
			t.Fatal(err)
		}
		requireExampleServer(t, response)
	})

	t.Run("panel connect json", func(t *testing.T) {
		client := platformconnect.NewPlatformServiceClient(server.Client(), server.URL, connect.WithProtoJSON())
		response, err := client.GetInfo(ctx, &platform.GetInfoRequest{})
		if err != nil {
			t.Fatal(err)
		}
		requireExampleServer(t, response)
	})
}
