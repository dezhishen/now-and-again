// Package notifier implements built-in notification channels.
//
// Adding a new channel:
//  1. Create a new file in this package (e.g., telegram.go).
//  2. Define a struct implementing the Notifier interface.
//  3. Call RegisterNotifier in an init() function.
//
// Example (notional):
//
//	package notifier
//	import "github.com/dezhishen/now-and-again/backend/internal/service"
//
//	type EmailNotifier struct { /* SMTP config */ }
//	func (n *EmailNotifier) Code() string { return "email" }
//	func (n *EmailNotifier) Send(ctx context.Context, msg *service.Message) error { ... }
//
//	func init() {
//	    service.RegisterNotifier(&EmailNotifier{})
//	}
//
// This open/closed design means new channels never require changes
// to notification_engine.go or any other core file.
package notifier
