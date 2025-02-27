package registration

import (
	"app/internal/domain/user"
	"context"
)

type Provider interface {
	Registration(user user.User) (user.User, error)
}
type OAuth struct {
	ctx            context.Context
	clientProvider Provider
}

func New(
	ctx context.Context,
	clientProvider Provider,
) *OAuth {
	return &OAuth{
		ctx:            ctx,
		clientProvider: clientProvider,
	}
}

//func (o *OAuth) Register(user user.User) (user.User, error) {
//	const op = "grpc-server.service.registration.register"
//	logging.L(o.ctx).Info("op", op)
//	ouath, _ := o.clientProvider.Registration(user)
//	return user.User{}, nil
//}
