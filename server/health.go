package server

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type HealthServer struct{}

// Check 标准grpc 健康检查
func (h *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	health := true
	//TODO 健康检查
	if !health {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
	}
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

// Watch 标准grpc 健康检查
func (h *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	resp := new(grpc_health_v1.HealthCheckResponse)
	health := true
	//TODO 健康检查
	if !health {
		resp.Status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	} else {
		resp.Status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	for range time.NewTicker(time.Second).C {
		err := stream.Send(resp)
		if err != nil {
			return status.Error(codes.Canceled, "Stream has ended.")
		}
	}
	return nil
}
