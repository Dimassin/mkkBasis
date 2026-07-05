package ports

import "context"

type EmailService interface {
	SendTeamInvite(ctx context.Context, email string, teamID, inviterID int) error
}
