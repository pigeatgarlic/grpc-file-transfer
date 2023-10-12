package server

import (
	"crypto/md5"
	"fmt"
	"os"
	"path"
	"time"

	mlspb "github.com/pigeatgarlic/grpc-file-server/pkg/protobuf"
	"google.golang.org/grpc"
)

// MLSServer
type MLSServer struct {
	mlspb.UnimplementedMLSServiceServer
	documentDir string
}

// NewMLSServer
func NewMLSServer(documentDir string) mlspb.MLSServiceServer {
	return &MLSServer{
		documentDir: documentDir,
	}
}

// RegisterMLSServer
func RegisterMLSServer(registrar grpc.ServiceRegistrar, server mlspb.MLSServiceServer) {
	mlspb.RegisterMLSServiceServer(registrar, server)
}

// Upload will receive a file and store it
func (ms *MLSServer) Upload(stream mlspb.MLSService_UploadServer) error {
	md, err := expandMetaData(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to retrieve incoming metadata: %s", err.Error())
	}

	size := md.fileSize
	tempFile := path.Join(ms.documentDir, fmt.Sprintf("%s.%d.temp", md.fileName, time.Now().UnixMilli()))
	destFile := path.Join(ms.documentDir, md.fileName)

	fmt.Printf("Begin receiving file %s\n", tempFile)
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %s", tempFile, err.Error())
	}
	done := false
	ids := []int64{}
	bytes_received := int64(0)
	channel := make(chan *mlspb.Chunk, 1000)
	defer func() {
		done = true
		for {
			time.Sleep(10 * time.Millisecond)
			stat, err := f.Stat()
			if err != nil {
				return
			} else if stat.Size() == bytes_received {
				break
			}
		}

		f.Close()
		if bytes_received < md.fileSize {
			fmt.Printf("Only received %dMB out of %dMB for file %s\n", bytes_received/1024/1024, size/1024/1024, tempFile)
			os.Remove(tempFile)
		} else {
			fmt.Printf("Finished receiving file %s\n", destFile)
			os.Remove(destFile)
			os.Rename(tempFile, destFile)
		}
	}()

	go func() {
		last_time := time.Now().UnixMilli()
		var last_received int64 = 0
		for {
			time.Sleep(5 * time.Second)
			if done {
				return
			}

			diff := time.Now().UnixMilli() - last_time
			bytes_diff := bytes_received - last_received
			last_received = bytes_received
			last_time = last_time + diff

			fmt.Printf("%dMB speed=%dMB\n", bytes_received/1024/1024, bytes_diff/1024/1024*1000/diff)
			if bytes_received > 0 && bytes_diff == 0 {
				fmt.Printf("no new bytes received, context canceled\n")
				stream.Context().Done()
			}
		}
	}()
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			err := stream.Send(&mlspb.UploadStatus{Success: ids})
			if err != nil {
				break
			}
		}
	}()
	go func() {
		var last_id int64 = 0
		for {
			data := <-channel
			if data == nil {
				return
			}

			if data.Sum256 != fmt.Sprintf("%x", md5.Sum(data.Content)) {
				panic(fmt.Errorf("invalid checksum"))
			} else if last_id != data.Id-1 {
				channel <- data
				continue
			}

			_, err = f.Write(data.Content)
			bytes_received += int64(len(data.Content))
			last_id = data.Id
			ids = append(ids, data.Id)
			if err != nil && !done {
				panic(fmt.Errorf("error Writing data: %s", err.Error()))
			}
		}
	}()
	for {
		data, err := stream.Recv()
		if err != nil {
			channel <- nil
			break
		}
		channel <- data
	}

	return nil
}
