package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/user"
	"github.com/cmrd-a/GophKeeper/server/insecure"
)

func main() {
	log.Println("its a client")
	get()
}

func get() {
	creds := credentials.NewClientTLSFromCert(insecure.CertPool, "localhost:8082")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.NewClient("localhost:8082", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := user.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := client.Register(ctx, &user.RegisterRequest{Login: "user", Password: "password"})
	if err != nil {
		log.Fatalf("client failed: %v", err)
	}
	log.Println(res)
}
