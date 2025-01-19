package rabbitmq

import (
	"app/internal/config"
	"app/pkg/common/logging"
	"app/pkg/utils/loop"
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type App struct {
	ctx  context.Context
	cfg  *config.Config
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New(
	ctx context.Context,
	cfg *config.Config,
) *App {
	return &App{
		ctx: ctx,
		cfg: cfg,
	}
}

func (a *App) Connect() error {

	logging.L(a.ctx).Info("Connecting to RabbitMQ")

	dns := fmt.Sprintf("amqp://%s:%s@%s:%d/", a.cfg.Queue.User, a.cfg.Queue.Pass, a.cfg.Queue.Host, a.cfg.Queue.Port)

	var conn *amqp.Connection
	var ch *amqp.Channel
	var err error

	err = loop.DoWithAttempt(a.ctx, func() error {
		conn, err = amqp.Dial(dns)
		if err != nil {
			logging.L(a.ctx).Error("Error connect to RabbitMQ: %v. Repeat after %s seconds...", err, a.cfg.Queue.MaxDelay)
			return err
		}
		ch, err = conn.Channel()
		if err != nil {
			logging.L(a.ctx).Error("Error opening the RabbitMQ channel: %v. Repeat after 5 seconds...", err)
			return err
		}
		return nil
	}, a.cfg.Queue.MaxAttempts, a.cfg.Queue.MaxDelay)

	logging.L(a.ctx).Info("RabbitMQ connected")

	a.conn = conn
	a.ch = ch

	return nil
}

func (a *App) SetupQueueAndExchange(NameQueue string) error {

	err := a.ch.ExchangeDeclare(
		"sso",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = a.ch.QueueDeclare(
		NameQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = a.ch.QueueBind(
		NameQueue,
		"",
		"sso",
		false,
		nil,
	)

	if err != nil {
		return err
	}

	return nil
}

func (a *App) Channel() *amqp.Channel {
	return a.ch
}

func (a *App) PublishMsg(msg []byte) {
	err := a.ch.Publish(
		"sso",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		},
	)
	if err != nil {
		logging.L(a.ctx).Error("Error when posting a message: %v", err)
	}
}

func (a *App) Close() {

	if a.conn == nil {
		_ = a.conn.Close()
	}

	if a.ch == nil {
		_ = a.ch.Close()
	}
}
