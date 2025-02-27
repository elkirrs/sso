package registration

import (
	"app/internal/domain/user"
	"app/pkg/grpc"
)

type OAuth interface {
	UserRegistration(user user.User) (user.User, error)
}

type serverGRPC struct {
	grpc.UnimplementedAuthServiceServer
	//oauth OAuth
}

//func Register(ctx context.Context, gRPC *grpc.Server, storage *pgxpool.Pool) {
//	storageClient, err := clientStorage.New(ctx, storage)
//	if err != nil {
//		logging.L(ctx).Error("failed to init storage client", err)
//		return
//	}
//	c := authSe.New(ctx, storageClient)
//	gRPCClient.RegisterAuthServiceServer(gRPC, &serverGRPC{oauth: c})
//}
