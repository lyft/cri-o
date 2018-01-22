package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/containers/storage"
	"github.com/docker/docker/pkg/mount"
	"golang.org/x/net/context"
	"golang.org/x/sys/unix"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
	"strings"
)

func getStorageFsInfo(store storage.Store) (*pb.FilesystemUsage, error) {
	rootPath := store.RunRoot()
	storageDriver := store.GraphDriverName()
	filePath := path.Join(rootPath, storageDriver)

	statInfo := syscall.Stat_t{}
	err := syscall.Lstat(filePath, &statInfo)
	if err != nil {
		return nil, err
	}

	mounts, err := mount.GetMounts()
	if err != nil {
		return nil, err
	}

	deviceName, err := getDeviceName(statInfo, mounts)
	if err != nil {
		return nil, err
	}

	uuid, err := getDeviceToUUID(deviceName)
	if err != nil {
		return nil, err
	}

	bytesUsed, inodesUsed := getDiskUsageStats(filePath)

	usage := pb.FilesystemUsage{
		Timestamp:  time.Now().UnixNano(),
		StorageId:  &pb.StorageIdentifier{uuid},
		UsedBytes:  &pb.UInt64Value{bytesUsed},
		InodesUsed: &pb.UInt64Value{inodesUsed},
	}

	return &usage, nil
}

func getDeviceToUUID (devicePath string) (string, error) {
	const dir = "/dev/disk/by-uuid"

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", nil
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		target, err := os.Readlink(path)
		if err != nil {
			continue
		}
		device, err := filepath.Abs(filepath.Join(dir, target))
		if err != nil {
			return "", fmt.Errorf("failed to resolve the absolute path of %q", filepath.Join(dir, target))
		}
		if strings.Compare(device, devicePath) == 0 {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("Device path not found")
}

func getDeviceName(info syscall.Stat_t, mounts []*mount.Info) (string, error) {
	queryMajor := int(unix.Major(uint64(info.Dev)))
	queryMinor := int(unix.Minor(uint64(info.Dev)))

	for _, mount := range mounts {
		if mount.Minor == queryMinor && mount.Major == queryMajor {
			return mount.Source, nil
		}
	}

	return "", fmt.Errorf("No match found")
}

func getDiskUsageStats(rootpath string) (uint64, uint64) {
	var dirSize uint64 = 0
	var inodeCount uint64 = 0

	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		fileStat, error := os.Lstat(rootpath)
		if error != nil {
			if fileStat.Mode()&os.ModeSymlink != 0 {
				// Is a symlink; no error should be returned
			}
			return error
		}

		dirSize += uint64(info.Size())
		inodeCount += 1

		return nil
	})

	if err != nil {
		fmt.Printf("walk error [%v]\n", err)
	}

	return dirSize, inodeCount
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (s *Server) ImageFsInfo(ctx context.Context, req *pb.ImageFsInfoRequest) (resp *pb.ImageFsInfoResponse, err error) {
	const operation = "image_fs_info"
	defer func() {
		recordOperation(operation, time.Now())
		recordError(operation, err)
	}()

	store := s.StorageImageServer().GetStore()
	fmt.Printf("%+v\n", store)
	fsUsage, err := getStorageFsInfo(store)

	return &pb.ImageFsInfoResponse{
		ImageFilesystems: []*pb.FilesystemUsage{fsUsage},
	}, nil

}
