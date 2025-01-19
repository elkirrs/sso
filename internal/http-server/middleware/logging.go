package middleware

import (
	"app/pkg/client/rabbitmq"
	"app/pkg/common/logging"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type RequestLog struct {
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	RemoteAddr string    `json:"remote_addr"`
	Timestamp  time.Time `json:"timestamp"`
	Request    string    `json:"request"`
	Response   string    `json:"response"`
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	StatusCode int
	Body       bytes.Buffer
}

func Logging(
	ctx context.Context,
	amqpClient *rabbitmq.App,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := NewResponseWriterWrapper(w)
			var requestContent string
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					logging.L(ctx).Error("Failed to read request body: %v", err)
				} else {
					requestContent = string(bodyBytes)
				}
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			next.ServeHTTP(wrappedWriter, r)

			logEntry := RequestLog{
				Method:     r.Method,
				Path:       r.URL.Path,
				RemoteAddr: r.RemoteAddr,
				Timestamp:  time.Now(),
				Request:    requestContent,
				Response:   wrappedWriter.Body.String(),
			}

			body, err := json.Marshal(logEntry)
			if err != nil {
				logging.L(ctx).Error("Error when serializing the log record: %v", err)
				next.ServeHTTP(wrappedWriter, r)
				return
			}

			go amqpClient.PublishMsg(body)
		})
	}
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: w,
		StatusCode:     http.StatusOK, // по умолчанию статус 200
	}
}

func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriterWrapper) Write(data []byte) (int, error) {
	rw.Body.Write(data)
	return rw.ResponseWriter.Write(data)
}
