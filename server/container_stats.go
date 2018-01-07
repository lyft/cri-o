package server

import (
	"fmt"
	"time"

	"github.com/kubernetes-incubator/cri-o/lib"
	"golang.org/x/net/context"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (s *Server) ContainerStats(ctx context.Context, req *pb.ContainerStatsRequest) (resp *pb.ContainerStatsResponse, err error) {
	const operation = "container_stats"
	defer func() {
		recordOperation(operation, time.Now())
		recordError(operation, err)
	}()

	container := s.GetContainer(req.ContainerId)
	if container == nil {
		return nil, fmt.Errorf("invalid container")
	}

	now := time.Now().UnixNano()
	stats, err := s.GetContainerStats(container, &lib.ContainerStats{})
	if err != nil {
		return nil, err
	}

	return &pb.ContainerStatsResponse{
		&pb.ContainerStats{
			Attributes: &pb.ContainerAttributes{
				Id:          req.ContainerId,
				Metadata:    container.Metadata(),
				Labels:      container.Labels(),
				Annotations: container.Annotations(),
			},
			Cpu: &pb.CpuUsage{
				Timestamp:            now,
				UsageCoreNanoSeconds: &pb.UInt64Value{stats.CPUNano + stats.SystemNano},
			},
			Memory: &pb.MemoryUsage{
				Timestamp:       now,
				WorkingSetBytes: &pb.UInt64Value{stats.MemUsage},
			},
			WritableLayer: nil,
		},
	}, nil
}
