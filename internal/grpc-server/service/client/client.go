package client

import (
	clientDomain "app/internal/domain/client"
	"app/pkg/common/logging"
	"context"
)

type Client struct {
	ctx            context.Context
	clientProvider Provider
}

type Provider interface {
	GetClientByName(name string) (clientDomain.Client, error)
}

func New(
	ctx context.Context,
	clientProvider Provider,
) *Client {
	return &Client{
		ctx:            ctx,
		clientProvider: clientProvider,
	}
}

func (c *Client) GetClientByName(name string) (clientDomain.Client, error) {
	const op = "grpc-server.service.client.GetClientByName"
	logging.L(c.ctx).Info("op", op)

	client, _ := c.clientProvider.GetClientByName(name)
	return clientDomain.Client{
		ID:     client.ID,
		Name:   client.Name,
		Secret: client.Secret,
	}, nil
}
