package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// APIConfig is the configuration for the API.
type APIConfig struct {
	Logger *logrus.Logger
}

// Router wraps the julienschmidt httprouter to add context and custom error handling.
type Router interface {
	http.Handler
}

type _router struct {
	router     *httprouter.Router
	baseLogger *logrus.Logger
}

// New creates a new API Router.
func New(cfg APIConfig) Router {
	rt := &_router{
		router:     httprouter.New(),
		baseLogger: cfg.Logger,
	}

	rt.registerRoutes()

	return rt
}

// registerRoutes collega i path HTTP alle funzioni handler Go.
func (rt *_router) registerRoutes() {

	// ---------------- AUTENTICAZIONE E SESSIONE ----------------
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// ---------------- GESTIONE UTENTE LOGGATO ----------------
	rt.router.PATCH("/utenti/me/username", rt.wrap(rt.setMyUserName))
	rt.router.PATCH("/utenti/me/photo", rt.wrap(rt.setMyPhoto))

	// ---------------- CONVERSAZIONI E GRUPPI ----------------
	rt.router.GET("/conversazioni", rt.wrap(rt.getMyConversations))
	rt.router.GET("/conversazioni/:conversazioneID", rt.wrap(rt.getConversation))
	rt.router.PATCH("/conversazioni/:conversazioneID/name", rt.wrap(rt.setGroupName))
	rt.router.PATCH("/conversazioni/:conversazioneID/photo", rt.wrap(rt.setGroupPhoto))

	// Partecipanti al gruppo
	rt.router.POST("/conversazioni/:conversazioneID/partecipanti", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/conversazioni/:conversazioneID/partecipanti/me", rt.wrap(rt.leaveGroup))

	// ---------------- MESSAGGI E COMMENTI ----------------
	rt.router.POST("/conversazioni/:conversazioneID/messaggi", rt.wrap(rt.sendMessage))
	rt.router.POST("/conversazioni/:conversazioneID/messaggi/forward", rt.wrap(rt.forwardMessage))
	rt.router.DELETE("/conversazioni/:conversazioneID/messaggi/:messaggioID", rt.wrap(rt.deleteMessage))

	// Commenti/Reazioni
	rt.router.POST("/conversazioni/:conversazioneID/messaggi/:messaggioID/commenti", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversazioni/:conversazioneID/messaggi/:messaggioID/commenti/:commentoID", rt.wrap(rt.uncommentMessage))
}

// ServeHTTP makes the router implement the http.Handler interface.
func (rt *_router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.router.ServeHTTP(w, r)
}
