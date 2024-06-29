package client

import (
	"app/internal/domain/client"
	clientService "app/internal/grpc-server/service/client"
	clientStorage "app/internal/storage/pgsql/client"
	"app/pkg/common/logging"
	gRPCClient "app/pkg/grpc"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	GetClientByName(name string) (client.Client, error)
}

type serverGRPC struct {
	//sso.UnimplementedAuthServer
	client Client
}

func Register(ctx context.Context, gRPC *grpc.Server, storage *pgxpool.Pool) {
	storageClient, err := clientStorage.New(ctx, storage)
	if err != nil {
		logging.L(ctx).Error("failed to init storage client", err)
		return
	}
	c := clientService.New(ctx, storageClient)
	gRPCClient.RegisterClientServiceServer(gRPC, &serverGRPC{client: c})
}

func (s *serverGRPC) GetClientSecret(
	ctx context.Context,
	req *gRPCClient.ClientRequest,
) (*gRPCClient.ClientResponse, error) {
	const op = "todo"
	logging.L(ctx).Info("op", op)

	if err := validationRequestClient(req); err != nil {
		return &gRPCClient.ClientResponse{}, err
	}

	clientData, err := s.client.GetClientByName(req.Name)
	if err != nil {
		return &gRPCClient.ClientResponse{}, err
	}
	return &gRPCClient.ClientResponse{
		Id:     clientData.ID,
		Name:   clientData.Name,
		Secret: clientData.Secret,
	}, nil
}

func validationRequestClient(req *gRPCClient.ClientRequest) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "invalid app name")
	}
	return nil
}
