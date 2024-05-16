package grpcProtocol

import (
	"context"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"

	pb "github.com/Nchezhegova/metrics-alerts/cmd/grpcProtocol/proto"
)

func StartGRPCClient() *pb.DataServiceClient {
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	c := pb.NewDataServiceClient(conn)
	return &c
}

func TestSend(c pb.DataServiceClient, data []byte) {
	md := metadata.New(map[string]string{
		"Content-Encoding": "gzip",
		"X-Real-IP":        config.IP,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := c.SendData(ctx, &pb.DataRequest{Message: data})
	if err != nil {
		log.Fatal(err)
	}
}
