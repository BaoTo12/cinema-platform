package async

import (
	"context"
	"time"

	"cinemaos-backend/internal/pkg/logger"
	"cinemaos-backend/internal/pkg/worker"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JobType represents predefined job types
type JobType string

const (
	JobTypeEmail       JobType = "email"
	JobTypeNotification JobType = "notification"
	JobTypeCleanup     JobType = "cleanup"
	JobTypeReport      JobType = "report"
)

// EmailPayload represents data for sending an email
type EmailPayload struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// NotificationPayload represents data for a notification
type NotificationPayload struct {
	UserID  uuid.UUID
	Type    string
	Title   string
	Message string
	Data    map[string]interface{}
}

// Dispatcher manages async job dispatching
type Dispatcher struct {
	pool   *worker.Pool
	logger *logger.Logger
}

// NewDispatcher creates a new async dispatcher
func NewDispatcher(workers, queueSize int, log *logger.Logger) *Dispatcher {
	pool := worker.NewPool("async-jobs", workers, queueSize, log)
	return &Dispatcher{
		pool:   pool,
		logger: log,
	}
}

// Start starts the dispatcher
func (d *Dispatcher) Start() {
	d.pool.Start()
	d.logger.Info("async dispatcher started")
}

// Stop stops the dispatcher gracefully
func (d *Dispatcher) Stop(timeout time.Duration) error {
	return d.pool.Stop(timeout)
}

// SubmitEmail submits an email job for async processing
func (d *Dispatcher) SubmitEmail(payload EmailPayload) bool {
	job := worker.Job{
		ID:      uuid.New().String(),
		Type:    string(JobTypeEmail),
		Payload: payload,
		Handler: d.handleEmail,
	}
	return d.pool.Submit(job)
}

// SubmitNotification submits a notification job
func (d *Dispatcher) SubmitNotification(payload NotificationPayload) bool {
	job := worker.Job{
		ID:      uuid.New().String(),
		Type:    string(JobTypeNotification),
		Payload: payload,
		Handler: d.handleNotification,
	}
	return d.pool.Submit(job)
}

// SubmitCleanup submits a cleanup job
func (d *Dispatcher) SubmitCleanup(cleanupFn func(ctx context.Context) error) bool {
	job := worker.Job{
		ID:   uuid.New().String(),
		Type: string(JobTypeCleanup),
		Handler: func(ctx context.Context, _ interface{}) error {
			return cleanupFn(ctx)
		},
	}
	return d.pool.Submit(job)
}

// handleEmail processes email jobs
func (d *Dispatcher) handleEmail(ctx context.Context, payload interface{}) error {
	email, ok := payload.(EmailPayload)
	if !ok {
		d.logger.Error("invalid email payload type")
		return nil // Don't retry on type error
	}

	d.logger.Info("sending email",
		zap.Strings("to", email.To),
		zap.String("subject", email.Subject),
	)

	// TODO: Integrate with email service (SendGrid, AWS SES, etc.)
	// For now, just simulate sending
	time.Sleep(100 * time.Millisecond)

	d.logger.Info("email sent successfully")
	return nil
}

// handleNotification processes notification jobs
func (d *Dispatcher) handleNotification(ctx context.Context, payload interface{}) error {
	notif, ok := payload.(NotificationPayload)
	if !ok {
		d.logger.Error("invalid notification payload type")
		return nil
	}

	d.logger.Info("sending notification",
		zap.String("user_id", notif.UserID.String()),
		zap.String("type", notif.Type),
		zap.String("title", notif.Title),
	)

	// TODO: Integrate with notification service (FCM, WebSockets, etc.)
	time.Sleep(50 * time.Millisecond)

	d.logger.Info("notification sent successfully")
	return nil
}

// QueueSize returns current pending jobs
func (d *Dispatcher) QueueSize() int {
	return d.pool.QueueSize()
}

// IsRunning returns if dispatcher is running
func (d *Dispatcher) IsRunning() bool {
	return d.pool.IsRunning()
}
