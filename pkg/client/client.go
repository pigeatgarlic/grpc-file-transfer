package client

import (
	"context"
	"io"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	mlspb "bitbucket.org/maxex/mls_poc/pkg/protobuf"
)

// MLSClient maintains info for talking to MLS service
type MLSClient struct {
	logger *logrus.Logger
	conn   *grpc.ClientConn
	route  mlspb.MLSServiceClient
}

// NewClient will create a New MLSClient
func NewClient(logger *logrus.Logger, conn *grpc.ClientConn) *MLSClient {
	return &MLSClient{
		logger: logger,
		conn:   conn,
		route:  mlspb.NewMLSServiceClient(conn),
	}
}

// Upload a file to the MLS Service
func (mc *MLSClient) Upload(ctx context.Context, f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}

	md := metadata.Pairs("file_name", info.Name(), "file_size", strconv.FormatInt(info.Size(), 10))
	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := mc.route.Upload(ctx)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024*16)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		if err := stream.Send(&mlspb.Chunk{Content: buf[:n]}); err != nil {
			if err == io.EOF {
				break
			}

			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	mc.logger.Infof("Finished sending file...")
	return nil
}

// CloseConn
func (mc *MLSClient) CloseConn() error {
	return mc.conn.Close()
}
