package server

import (
	"crypto/md5"
	"fmt"
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
	md, err := expandMetaData(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to retrieve incoming metadata: %s", err.Error())
	}

	size := md.fileSize
	tempFile := path.Join(ms.documentDir,fmt.Sprintf("%s.%d.temp",md.fileName,time.Now().UnixMilli()))
	destFile := path.Join(ms.documentDir,md.fileName)

	ms.logger.Infof("Begin receiving file %s",tempFile)
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
		f.Close()
		if bytes_received < md.fileSize {
			ms.logger.Errorf("Only received %d bytes out of the expect %d bytes for file %s", bytes_received, size, tempFile)
			os.Remove(tempFile)
		} else {
			ms.logger.Infof("Finished receiving file %s", destFile)
			os.Remove(destFile)
			os.Rename(tempFile,destFile)
		}
	}()



	go func()  {
		last := time.Now().UnixMilli()
		for {
			time.Sleep(1 * time.Second)
			stat,err := f.Stat()
			if err != nil {
				return
			}

			bytes_received = stat.Size()
			total := ( time.Now().UnixMilli() - last ) / 1000
			speed := stat.Size() / 1024 / 1024 / (total)
			ms.logger.Infof("%dMB speed=%dMB/s\n",stat.Size() / 1024 / 1024,speed)
		}
	}()
	go func ()  {
		for {
			time.Sleep(time.Second)
			err := stream.Send(&mlspb.UploadStatus{ Success: ids, })
			if err != nil {
				break
			}
		}
	}()
	go func ()  {
		var last_id int64 = 0
		for {
			data := <-channel
			if data == nil {
				return
			}

			if data.Sum256 != fmt.Sprintf("%x",md5.Sum(data.Content)) {
				panic(fmt.Errorf("invalid checksum"))
			} else if last_id != data.Id - 1 {
				channel<-data
				continue
			}

			last_id = data.Id
			ids = append(ids, data.Id)
			_,err = f.Write(data.Content)
			if err != nil && !done {
				panic(fmt.Errorf("error Writing data: %s", err.Error()))
			}
		}
	}()
	for {
		data, err := stream.Recv()
		if err != nil { 
			channel<-nil
			break 
		}
		channel<-data
	}

	return nil
}
