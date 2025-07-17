package rabbitmq

import (
	"app/internal/config"
	"app/internal/queue"
	"app/pkg/common/logging"
	"app/pkg/utils/loop"
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type App struct {
	ctx  context.Context
	cfg  config.Queue
	conn *amqp.Connection
	ch   *amqp.Channel
}

type MessageProcessor interface {
	Process(ctx context.Context, msg amqp.Delivery) error
}

func New(
	ctx context.Context,
	cfg config.Queue,
) (*App, error) {
	app := &App{
		ctx: ctx,
		cfg: cfg,
	}

	if cfg.Driver != "" {
		err := app.Connect()
		if err != nil {
			logging.L(ctx).Error("Couldn't connect to RabbitMQ: ", err)
			return nil, err
		}

		for _, queueData := range queue.List {
			err = app.SetupQueueAndExchange(queueData.Exchange, queueData.Queue, queueData.RoutingKey)
			if err != nil {
				logging.L(ctx).Error("Error during setup RabbitMQ: ", err)
				return nil, err
			}
		}
	}

	return app, nil
}

func (a *App) Connect() error {

	logging.L(a.ctx).Info("Connecting to RabbitMQ")

	dns := fmt.Sprintf("amqp://%s:%s@%s:%d/", a.cfg.User, a.cfg.Pass, a.cfg.Host, a.cfg.Port)

	var conn *amqp.Connection
	var ch *amqp.Channel
	var err error

	err = loop.DoWithAttempt(a.ctx, func() error {
		conn, err = amqp.Dial(dns)
		if err != nil {
			logging.L(a.ctx).Error("Error connect to RabbitMQ: %v. Repeat after %s seconds...", err, a.cfg.MaxDelay)
			return err
		}
		ch, err = conn.Channel()
		if err != nil {
			logging.L(a.ctx).Error("Error opening the RabbitMQ channel: %v. Repeat after 5 seconds...", err)
			return err
		}
		return nil
	}, a.cfg.MaxAttempts, a.cfg.MaxDelay)

	logging.L(a.ctx).Info("RabbitMQ connected")

	a.conn = conn
	a.ch = ch

	return nil
}

func (a *App) SetupQueueAndExchange(exchangeName, queueName, routingKey string) error {

	err := a.ch.ExchangeDeclare(
		exchangeName,
		"direct",
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
		queueName,
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
		queueName,
		routingKey,
		exchangeName,
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

func (a *App) PublishMsg(exchangeName, routingKey string, msg []byte) {
	err := a.ch.Publish(
		exchangeName,
		routingKey,
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

func (a *App) ConsumeMsg(
	queueName string,
	handler func(amqp.Delivery),
) {
	ch, err := a.conn.Channel()

	if err != nil {
		logging.L(a.ctx).Error("failed to open a channel", err)
		return
	}

	defer ch.Close()

	messages, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logging.L(a.ctx).Error(fmt.Sprintf("failed to register consumer for %s", queueName), err)
		return
	}

	for msg := range messages {
		logging.L(a.ctx).Info(fmt.Sprintf("Received message from %s: %s", queueName, string(msg.Body)))
		handler(msg)
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

func ProcessMessage(
	ctx context.Context,
	msg amqp.Delivery,
	processor MessageProcessor,
) {
	err := processor.Process(ctx, msg)

	if err != nil {
		logging.L(ctx).Error("Failed to process message", "error", err, "message", string(msg.Body))

		if isRetryableError(err) {
			err := msg.Nack(false, true)
			if err != nil {
				return
			}
		} else {
			err := msg.Reject(false)
			if err != nil {
				return
			}
		}
		return
	}

	err = msg.Ack(false)
	if err != nil {
		return
	}
}

func isRetryableError(err error) bool {
	return err.Error() == "temporary failure" || err.Error() == "database unavailable"
}
