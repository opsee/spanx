// +build ignore

package main

import (
	"crypto/tls"
	"fmt"

	"github.com/opsee/basic/schema"
	svc "github.com/opsee/basic/service"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	req := &svc.EnhancedCombatModeRequest{
		User: &schema.User{
			Id:         int32(1),
			CustomerId: "ddaadd1a-78e8-11e5-856d-b79e2da78aac",
			Email:      "greg@opsee.com",
			Name:       "Greg Poirier",
			Verified:   true,
			Admin:      true,
			Active:     true,
		},
		Region: "us-west-1",
	}

	conn, err := grpc.Dial("spanx.in.opsee.com:8443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		panic(err)
	}
	client := svc.NewSpanxClient(conn)

	resp, err := client.EnhancedCombatMode(context.Background(), req)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.StackUrl)
}
