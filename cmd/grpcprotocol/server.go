package grpcprotocol

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	pb "github.com/Nchezhegova/metrics-alerts/cmd/grpcprotocol/proto"
)

type DataServiceServer struct {
	pb.UnimplementedDataServiceServer
	m             storage.MStorage
	trustedSubnet string
	mu            sync.Mutex
}

func (s *DataServiceServer) SendData(ctx context.Context, in *pb.DataRequest) (*pb.DataResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
		s.m.GaugeStorage(ctx, k, *v)

	case config.Counter:
		k := metrics.ID
		v := metrics.Delta
		s.m.CountStorage(ctx, k, *v)
		vNew, _ := s.m.GetCount(ctx, metrics.ID)
		metrics.Delta = &vNew

	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}

	return &pb.DataResponse{}, nil
}

func (s *DataServiceServer) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch in.Type {
	case config.Gauge:
		k := in.Name
		v, err := strconv.ParseFloat(in.Value, 64)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid gauge value")
		}
		s.m.GaugeStorage(ctx, k, v)
	case config.Counter:
		k := in.Name
		v, err := strconv.ParseInt(in.Value, 10, 64)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid counter value")
		}
		s.m.CountStorage(ctx, k, v)
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}
	return &pb.UpdateMetricResponse{}, nil
}

func (s *DataServiceServer) UpdateBatchMetrics(ctx context.Context, in *pb.UpdateBatchMetricsRequest) (*pb.UpdateBatchMetricsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metricsList []storage.Metrics
	var b io.ReadCloser

	if strings.Contains(in.ContentEncoding, "gzip") {
		gz, err := gzip.NewReader(bytes.NewReader([]byte(in.Message)))
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		b = gz
	} else {
		b = io.NopCloser(bytes.NewReader([]byte(in.Message)))
	}

	decoder := json.NewDecoder(b)
	err := decoder.Decode(&metricsList)
	if err != nil {
		return nil, err
	}

	for _, metrics := range metricsList {
		switch metrics.MType {
		case config.Gauge:
			k := metrics.ID
			v := metrics.Value
			s.m.GaugeStorage(ctx, k, *v)
		case config.Counter:
			k := metrics.ID
			v := metrics.Delta
			s.m.CountStorage(ctx, k, *v)
		default:
			return nil, status.Error(codes.InvalidArgument, "unknown metric type")
		}
	}

	return &pb.UpdateBatchMetricsResponse{}, nil
}

func (s *DataServiceServer) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	switch in.Type {
	case config.Gauge:
		v, exists := s.m.GetGauge(ctx, in.Name)
		if !exists {
			return nil, status.Error(codes.NotFound, "gauge not found")
		}
		return &pb.GetMetricResponse{Value: &pb.GetMetricResponse_GaugeValue{GaugeValue: v}}, nil
	case config.Counter:
		v, exists := s.m.GetCount(ctx, in.Name)
		if !exists {
			return nil, status.Error(codes.NotFound, "counter not found")
		}
		return &pb.GetMetricResponse{Value: &pb.GetMetricResponse_CounterValue{CounterValue: v}}, nil
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown metric type")
	}
}

func unaryInterceptor(s *DataServiceServer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
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
		ip := net.ParseIP(ipStr)
		_, ipNet, err := net.ParseCIDR(s.trustedSubnet)
		if err != nil {
			return nil, status.Error(codes.Internal, "error parsing trusted subnet")
		}
		if !ipNet.Contains(ip) {
			return nil, status.Error(codes.Unauthenticated, "IP is not in the trusted subnet")
		}
		return handler(ctx, req)
	}
}

func StartGRPCServer(memory storage.MStorage, trustedSubnet string) {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	server := &DataServiceServer{
		m:             memory,
		trustedSubnet: trustedSubnet,
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor(server)))
	pb.RegisterDataServiceServer(s, server)

	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
