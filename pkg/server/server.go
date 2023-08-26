package server

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	mlspb "github.com/pigeatgarlic/grpc-file-server/pkg/protobuf"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

// MLSServer
type MLSServer struct {
	mlspb.UnimplementedMLSServiceServer
	logger      *log.Logger
	documentDir string
}

// NewMLSServer
func NewMLSServer(logger *log.Logger, documentDir string) mlspb.MLSServiceServer {
	return &MLSServer{
		logger:      logger,
		documentDir: documentDir,
	}
}

// RegisterMLSServer
func RegisterMLSServer(registrar grpc.ServiceRegistrar, server mlspb.MLSServiceServer) {
	mlspb.RegisterMLSServiceServer(registrar, server)
}

// Upload will receive a file and store it
func (ms *MLSServer) Upload(stream mlspb.MLSService_UploadServer) error {
	ms.logger.Info("Begin receiving file...")

	md, err := expandMetaData(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to retrieve incoming metadata: %s", err.Error())
	}

	pathOfData := path.Join(ms.documentDir, md.fileName)
	f, err := os.OpenFile(pathOfData, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %s", md.fileName, err.Error())
	}
	defer f.Close()

	totalBytesReceived := int64(0)

	done := false
	go func()  {
		last := time.Now().UnixMilli()
		for {
			if done {
				return
			}

			time.Sleep(1 * time.Second)
			stat,err := f.Stat()
			if err != nil {
				return
			}

			total := ( time.Now().UnixMilli() - last ) / 1000
			speed := stat.Size() / 1024 / 1024 / (total)
			fmt.Printf("%dMB speed=%dMB/s\n",stat.Size() / 1024 / 1024,speed)
		}
	}()

	for {
		data, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			ms.logger.Errorf("error receiving data: %s", err.Error())
			break
		}

		totalBytesReceived += int64(len(data.GetContent()))

		_, err = f.Write(data.GetContent())
		if err != nil {
			ms.logger.Errorf("error Writing data: %s", err.Error())
			break
		}
	}

	if totalBytesReceived < md.fileSize {
		ms.logger.Errorf("Only received %d bytes out of the expect %d bytes for file %s", totalBytesReceived, md.fileSize, md.fileName)
		defer os.Remove(pathOfData)
	}

	ms.logger.Infof("Finished receiving file %s", md.fileName)
	done = true

	return nil
}
