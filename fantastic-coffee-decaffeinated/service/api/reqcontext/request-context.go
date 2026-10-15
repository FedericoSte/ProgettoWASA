/*
Package reqcontext contains the request context. Each request will have its own instance of RequestContext filled by the
middleware code in the api-context-wrapper.go (parent package).

Each value here should be assumed valid only per request only, with some exceptions like the logger.
*/
package reqcontext

import (
	"github.com/sirupsen/logrus"
)

// RequestContext is the context of the request, for request-dependent parameters

// RequestContext is the context of the current HTTP request.
// It contains the request ID and, se autenticato, l'ID dell'utente.
type RequestContext struct {
	ReqUUID string

	UserID string // <- AGGIUNTO: Conterrà l'ID dell'utente se il Bearer token è valido

	// Logger is a custom field logger for the request
	Logger logrus.FieldLogger
}
