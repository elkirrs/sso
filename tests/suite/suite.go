package suite

//
//import (
//	"app/internal/config"
//	sso "app/pkg/http/grpc-server"
//	"context"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials/insecure"
//	"net"
//	"os"
//	"strconv"
//	"testing"
//)
//
//type Suite struct {
//	*testing.T
//	Cfg        *config.Config
//	AuthClient sso.AuthClient
//}
//
//const (
//	grpcHost = "localhost"
//)
//
//// New creates new tests suite.
////
//// TODO: for pipeline tests we need to wait for app is ready
//func New(t *testing.T) (context.Context, *Suite) {
//	t.Helper()
//	t.Parallel()
//
//	cfg := config.MustLoadPath(configPath())
//
//	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
//
//	t.Cleanup(func() {
//		t.Helper()
//		cancelCtx()
//	})
//
//	cc, err := grpc.DialContext(context.Background(),
//		grpcAddress(cfg),
//		grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		t.Fatalf("grpc-server server connection failed: %v", err)
//	}
//
//	return ctx, &Suite{
//		T:          t,
//		Cfg:        cfg,
//		AuthClient: sso.NewAuthClient(cc),
//	}
//}
//
//func configPath() string {
//	const key = "CONFIG_PATH"
//
//	if v := os.Getenv(key); v != "" {
//		return v
//	}
//
//	return "../config/local_tests.yaml"
//}
//
//func grpcAddress(cfg *config.Config) string {
//	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
//}
