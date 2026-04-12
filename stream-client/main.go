package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	orderv1 "github.com/mxkwnz/ap2-generated/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <order-id>")
	}
	orderID := os.Args[1]

	conn, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := orderv1.NewOrderServiceClient(conn)
	stream, err := client.SubscribeToOrderUpdates(
		context.Background(),
		&orderv1.OrderRequest{OrderId: orderID},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscribed to order:", orderID)
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[UPDATE] Status: %s | Message: %s\n",
			update.Status, update.Message)
	}
}
