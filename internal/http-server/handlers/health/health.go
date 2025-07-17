package health

import (
	"app/internal/config"
	"app/pkg/client/rabbitmq"
	resp "app/pkg/common/core/api/response"
	"app/pkg/common/logging"
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"net/http"
	"time"
)

type Response struct {
	Dependencies map[string]bool `json:"dependencies,omitempty"`
	Uptime       string          `json:"uptime,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
}

var startTime = time.Now()

func New(
	ctx context.Context,
	cfg *config.Config,
	pgClient *pgxpool.Pool,
	queueClient *rabbitmq.App,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.health.New"

		logging.L(ctx).With(
			logging.StringAttr("op", op),
			logging.StringAttr("request_id", middleware.GetReqID(r.Context())),
		)

		status := "ok"
		dependencies := make(map[string]bool)
		dependencies["database"] = databaseStatus(pgClient)

		if cfg.Queue.Driver != "" {
			dependencies["rabbitmq"] = rabbitMQStatus(queueClient.Channel())
		}

		for _, value := range dependencies {
			if !value {
				status = "fail"
				break
			}
		}

		var response = &Response{
			Dependencies: dependencies,
			Uptime:       time.Since(startTime).String(),
			Timestamp:    time.Now().UTC(),
		}

		w.Header().Set("Content-Type", "application/json")
		if status == "fail" {
			w.WriteHeader(http.StatusServiceUnavailable)
			resp.Error(w, r, response)
		} else {
			resp.Ok(w, r, response)
		}
		return
	}
}

func rabbitMQStatus(channel *amqp.Channel) bool {
	if channel == nil || channel.IsClosed() {
		return false
	}
	return true
}

func databaseStatus(pgClient *pgxpool.Pool) bool {
	err := pgClient.Ping(context.Background())
	return err == nil
}
