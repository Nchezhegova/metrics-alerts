package grpcprotocol

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	pb "github.com/Nchezhegova/metrics-alerts/cmd/grpcprotocol/proto"
)

var mu sync.Mutex
var m storage.MStorage
var trustedSubnet string

type DataServiceServer struct {
	pb.UnimplementedDataServiceServer
}

func (s *DataServiceServer) SendData(ctx context.Context, in *pb.DataRequest) (*pb.DataResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	var contentEncoding string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("Content-Encoding")
		if len(values) > 0 {
			contentEncoding = values[0]
		}
	}

	var metrics storage.Metrics
	var b io.ReadCloser

	if strings.Contains(contentEncoding, "gzip") {
		gz, err := gzip.NewReader(bytes.NewReader(in.Message))
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		b = gz
	} else {
		b = io.NopCloser(bytes.NewReader(in.Message))
	}

	decoder := json.NewDecoder(b)
	err := decoder.Decode(&metrics)
	if err != nil {
		return nil, err
	}

	switch metrics.MType {
	case config.Gauge:
		k := metrics.ID
		v := metrics.Value
		m.GaugeStorage(ctx, k, *v)

	case config.Counter:
		k := metrics.ID
		v := metrics.Delta
		m.CountStorage(ctx, k, *v)
		vNew, _ := m.GetCount(ctx, metrics.ID)
		metrics.Delta = &vNew

	default:
		return nil, fmt.Errorf("unknowning metric type")
	}

	return &pb.DataResponse{}, nil
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	var ipStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			ipStr = values[0]
		}
	}
	if len(ipStr) == 0 {
		return nil, status.Error(codes.NotFound, "missing token")
	}
	//if token != SecretToken {
	//	return nil, status.Error(codes.Unauthenticated, "invalid token")
	//}
	ip := net.ParseIP(ipStr)
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, status.Error(codes.Internal, "error parsing trusted subnet")
	}
	if !ipNet.Contains(ip) {
		return nil, status.Error(codes.Unauthenticated, "IP is not in the trusted subnet")
	}
	return handler(ctx, req)
}

func StartGRPCServer(memory storage.MStorage, t string) {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	m = memory
	trustedSubnet = t
	s := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor))
	pb.RegisterDataServiceServer(s, &DataServiceServer{})

	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
