package handlers

import (
	"app/internal/domain/user"
	"app/internal/queue"
	"app/internal/storage"
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type HandleRegistration struct {
	dbClient    *pgxpool.Pool
	queueClient *rabbitmq.App
	storages    *storage.Storage
}

func NewHandleRegistration(
	dbClient *pgxpool.Pool,
	queueClient *rabbitmq.App,
	storages *storage.Storage,
) *HandleRegistration {
	return &HandleRegistration{
		dbClient:    dbClient,
		queueClient: queueClient,
		storages:    storages,
	}
}

func (h *HandleRegistration) Process(
	ctx context.Context,
	msg amqp.Delivery,
) error {
	logging.L(ctx).Info("Processing user registration", "data", string(msg.Body))

	if string(msg.Body) == "error" {
		return fmt.Errorf("invalid registration data")
	}

	var userData user.CreateUser
	if err := json.Unmarshal(msg.Body, &userData); err != nil {
		logging.L(ctx).Error("failed to decode message body", "error", err, "body", string(msg.Body))
		return fmt.Errorf("failed to decode message: %w", err)
	}

	timeUnix := time.Now().Unix()
	userData.CreatedAt = timeUnix
	userData.UpdatedAt = timeUnix

	logging.L(ctx).Info("user", userData)
	if userData.Name == "" || userData.Email == "" || userData.Password == "" {
		err := fmt.Errorf("missing required user fields: Name or Email or Possword")
		logging.L(ctx).Error("invalid user data", "error", err, "user", userData)
		return err
	}

	if err := h.storages.User.Registration(&userData); err != nil {
		if storage.ErrorCode(err) == storage.ErrCodeExists {
			logging.L(ctx).Info("user already exists")
			return nil
		}

		logging.L(ctx).Error("failed to save user in database", "error", err, "user", userData)
		return fmt.Errorf("failed to save user: %w", err)
	}

	logging.L(ctx).Info("user registration successful", "user", userData)

	var userSignal user.CreateUserSignal
	userSignal.UUID = userData.UUID
	userSignal.Service = "sso"
	userSignal.CreatedAt = time.Now().Unix()

	userJSON, _ := json.Marshal(userSignal)
	h.queueClient.PublishMsg(
		queue.List["usrRegSignal"].Exchange,
		queue.List["usrRegSignal"].RoutingKey,
		userJSON,
	)

	return nil
}
