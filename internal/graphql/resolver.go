package graphql

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
import (
	"numerous.com/cli/internal/appdev"
)

type Resolver struct {
	AppSessionsRepo    appdev.AppSessionRepository
	ToolSessionService appdev.AppSessionService
}
