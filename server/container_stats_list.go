package server

import (
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// ListContainerStats returns stats of all running containers.
func (s *Server) ListContainerStats(ctx context.Context, req *pb.ListContainerStatsRequest) (resp *pb.ListContainerStatsResponse, err error) {
	const operation = "list_container_stats"
	defer func() {
		recordOperation(operation, time.Now())
		recordError(operation, err)
	}()

	// This is an inefficient method, since the container will be resolved twice,
	// once by the container list code and once by the GetContainerStats call.
	containers, err := s.ListContainers(ctx, &pb.ListContainersRequest{
		&pb.ContainerFilter{
			Id:            req.Filter.Id,
			PodSandboxId:  req.Filter.PodSandboxId,
			LabelSelector: req.Filter.LabelSelector,
		},
	})
	if err != nil {
		return nil, err
	}

	var allStats []*pb.ContainerStats

	for _, container := range containers.Containers {
		stats, err := s.ContainerStats(ctx, &pb.ContainerStatsRequest{ContainerId: container.Id})
		if err != nil {
			logrus.Warn("unable to get stats for container %s", container.Id)
			continue
		}
		allStats = append(allStats, stats.Stats)
	}

	return &pb.ListContainerStatsResponse{
		Stats: allStats,
	}, nil
}
