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
	hms_addr = flag.String("hms", "localhost:10000",
		"HMS endpoint")
	proxy_addr = flag.String("proxy", "localhost:8080", "Proxy endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterMetastoreHandlerFromEndpoint(ctx, mux, *hms_addr, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(*proxy_addr, mux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
