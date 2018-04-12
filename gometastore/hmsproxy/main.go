//go:generate protoc -I../../protobuf -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway ../../protobuf/metastore.proto --grpc-gateway_out=logtostderr=true:../protobuf

package main

import (
	"flag"
	"log"
	"net/http"

	gw "github.com/akolb1/hmsv2api/gometastore/protobuf"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	hmsAddr = flag.String("hms", "localhost:10010",
		"HMS endpoint")
	proxyAddr = flag.String("proxy", "localhost:8080", "Proxy endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterMetastoreHandlerFromEndpoint(ctx, mux, *hmsAddr, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(*proxyAddr, mux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
