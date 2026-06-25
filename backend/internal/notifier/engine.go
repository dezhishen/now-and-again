package notifier

import (
	"context"
	"log"

	"github.com/dezhishen/now-and-again/backend/internal/repository"
)

// NotificationEngine is the central dispatcher that:
// 1. Resolves which channels a user has enabled for a given trigger event.
// 2. Renders the appropriate template.
// 3. Calls the registered Notifier to send the message.
// 4. Persists a Notification delivery record.
//
// To add a new channel (e.g., Telegram):
//   1. Implement the Notifier interface.
//   2. Call notifier.Register("telegram", yourImpl).
//   3. INSERT a row into notification_channels.
// No changes needed in this file.

// Notifier is the interface every notification channel must implement.
type Notifier interface {
	Code() string
	Send(ctx context.Context, msg *Message) error
}

// Message is the canonical outbound payload.
type Message struct {
	To           string
	Title        string
	Body         string
	TaskID       string
	TriggerEvent string
}

// notifierRegistry holds all registered channel implementations.
var notifierRegistry = map[string]Notifier{}

// Register adds a Notifier implementation. Call from init() in channel packages.
func RegisterNotifier(n Notifier) {
	notifierRegistry[n.Code()] = n
	log.Printf("notifier registered: %s", n.Code())
}

type NotificationEngine struct {
	repo     *repository.NotificationRepo
	userRepo *repository.UserRepo
}

func NewNotificationEngine(repo *repository.NotificationRepo, userRepo *repository.UserRepo) *NotificationEngine {
	return &NotificationEngine{repo: repo, userRepo: userRepo}
}

// Dispatch sends a notification for a given task and trigger event to the specified user.
// It checks user preferences (quiet hours, enabled channels) before sending.
func (e *NotificationEngine) Dispatch(ctx context.Context, userID, taskID string, event string) error {
	// TODO: implement full dispatch logic
	// 1. Query user_channel_configs WHERE user_id=? AND is_enabled=true
	// 2. Check quiet hours
	// 3. Resolve template (family-level > system-level)
	// 4. Lookup notifierRegistry[channelCode]
	// 5. notifier.Send(ctx, msg)
	// 6. INSERT INTO notifications
	log.Printf("[notif] dispatch: user=%s task=%s event=%s", userID, taskID, event)
	return nil
}

// DispatchToAssignees sends to all assignees of a task.
func (e *NotificationEngine) DispatchToAssignees(ctx context.Context, taskID string, event string) error {
	// TODO: query task assignees, then Dispatch to each
	log.Printf("[notif] dispatch to assignees: task=%s event=%s", taskID, event)
	return nil
}
