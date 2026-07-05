package ports

import "context"

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, userID, email, username string) error
}
