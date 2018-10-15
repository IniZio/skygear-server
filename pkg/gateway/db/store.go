package db

import (
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// GatewayStore provide functions to query application info from config db
type GatewayStore interface {

	// GetAppByDomain fetches the App with domain
	GetAppByDomain(domain string, app *model.App) error

	Close() error
}
