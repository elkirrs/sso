package handlers

import (
	"app/internal/domain/user"
	"app/internal/storage"
	"app/pkg/common/logging"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type HandleRegistration struct {
	dbClient *pgxpool.Pool
	storages *storage.Storage
}

func NewHandleRegistration(
	dbClient *pgxpool.Pool,
	storages *storage.Storage,
) *HandleRegistration {
	return &HandleRegistration{
		dbClient: dbClient,
		storages: storages,
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

	var user user.CreateUser
	if err := json.Unmarshal(msg.Body, &user); err != nil {
		logging.L(ctx).Error("failed to decode message body", "error", err, "body", string(msg.Body))
		return fmt.Errorf("failed to decode message: %w", err)
	}

	timeUnix := time.Now().Unix()
	user.CreatedAt = timeUnix
	user.UpdatedAt = timeUnix

	logging.L(ctx).Info("user", user)
	if user.Name == "" || user.Email == "" || user.Password == "" {
		err := fmt.Errorf("missing required user fields: Name or Email or Possword")
		logging.L(ctx).Error("invalid user data", "error", err, "user", user)
		return err
	}

	if err := h.storages.User.Registration(&user); err != nil {
		if storage.ErrorCode(err) == storage.ErrCodeExists {
			logging.L(ctx).Info("user already exists")
			return nil
		}

		logging.L(ctx).Error("failed to save user in database", "error", err, "user", user)
		return fmt.Errorf("failed to save user: %w", err)
	}

	logging.L(ctx).Info("user registration successful", "user", user)
	return nil
}
