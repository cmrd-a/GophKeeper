package gateway

import (
	"context"

	"fmt"
	"log"

	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"

	thirdparty "github.com/cmrd-a/GophKeeper/gen"
	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"
	"github.com/cmrd-a/GophKeeper/server/insecure"

	"io/fs"
	"mime"

	"strings"

	"google.golang.org/grpc/credentials"
)

// getOpenAPIHandler serves an OpenAPI UI.
func getOpenAPIHandler() http.Handler {
	mime.AddExtensionType(".svg", "image/svg+xml")
	// Use subdirectory in embedded files
	subFS, err := fs.Sub(thirdparty.OpenAPI, "OpenAPI")
	if err != nil {
		panic("couldn't create sub filesystem: " + err.Error())
	}
	return http.FileServer(http.FS(subFS))
}

// Run runs the gRPC-Gateway, dialling the provided address.
func Run(dialAddr string, HTTPPort int16) error {
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.NewClient(
		"dns:///"+dialAddr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(insecure.CertPool, "")),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %w", err)
	}

	gwmux := runtime.NewServeMux()
	err = user.RegisterUserServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	err = vault.RegisterVaultServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	oa := getOpenAPIHandler()

	gatewayAddr := fmt.Sprintf("0.0.0.0:%d", HTTPPort)
	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				gwmux.ServeHTTP(w, r)
				return
			}
			oa.ServeHTTP(w, r)
		}),
	}
	// Empty parameters mean use the TLS Config specified with the server.
	// if strings.ToLower(os.Getenv("SERVE_HTTP")) == "true" {
	log.Println("Serving gRPC-Gateway and OpenAPI Documentation on http://", gatewayAddr)
	return fmt.Errorf("serving gRPC-Gateway server: %w", gwServer.ListenAndServe())
	// }

	// log.Println("Serving gRPC-Gateway and OpenAPI Documentation on https://", gatewayAddr)
	// return fmt.Errorf("serving gRPC-Gateway server: %w", gwServer.ListenAndServeTLS("", ""))
}
