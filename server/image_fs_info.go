package server

import (
	"time"

	"golang.org/x/net/context"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// ImageFsInfo returns information of the filesystem that is used to store images.
func (s *Server) ImageFsInfo(ctx context.Context, req *pb.ImageFsInfoRequest) (resp *pb.ImageFsInfoResponse, err error) {
	const operation = "image_fs_info"
	defer func() {
		recordOperation(operation, time.Now())
		recordError(operation, err)
	}()

	usage := pb.FilesystemUsage{
		Timestamp:  time.Now().UnixNano(),
		StorageId:  &pb.StorageIdentifier{"00000000-0000-0000-0000-000000000000"},
		UsedBytes:  &pb.UInt64Value{0},
		InodesUsed: &pb.UInt64Value{0},
	}
	return &pb.ImageFsInfoResponse{
		ImageFilesystems: []*pb.FilesystemUsage{&usage},
	}, nil
}
