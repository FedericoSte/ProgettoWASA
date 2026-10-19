package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// APIConfig is the configuration for the API.
type APIConfig struct {
	Logger   *logrus.Logger
	Database database.AppDatabase
}

// Router wraps the julienschmidt httprouter to add context and custom error handling.
type Router interface {
	http.Handler
}

type _router struct {
	router     *httprouter.Router
	baseLogger *logrus.Logger
	db         database.AppDatabase
}

// New creates a new API Router.
func New(cfg APIConfig) Router {
	rt := &_router{
		router:     httprouter.New(),
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	rt.registerRoutes()

	return rt
}

// registerRoutes collega i path HTTP alle funzioni handler Go.
func (rt *_router) registerRoutes() {

	// ---------------- AUTENTICAZIONE E SESSIONE (Senza Lucchetto) ----------------
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// ---------------- GESTIONE UTENTE LOGGATO (Con Lucchetto: wrapAuth) ----------------
	rt.router.PATCH("/utenti/me/username", rt.wrapAuth(rt.setMyUserName))
	rt.router.PATCH("/utenti/me/photo", rt.wrapAuth(rt.setMyPhoto))

	// ---------------- CONVERSAZIONI E GRUPPI ----------------
	rt.router.GET("/conversazioni", rt.wrapAuth(rt.getMyConversations))

	// NUOVA ROTTA PER LA CREAZIONE DELLA CONVERSAZIONE
	rt.router.POST("/conversazioni", rt.wrapAuth(rt.createConversation))

	rt.router.GET("/conversazioni/:conversazioneID", rt.wrapAuth(rt.getConversation))
	rt.router.PATCH("/conversazioni/:conversazioneID/name", rt.wrapAuth(rt.setGroupName))
	rt.router.PATCH("/conversazioni/:conversazioneID/photo", rt.wrapAuth(rt.setGroupPhoto))

	// Partecipanti al gruppo
	rt.router.POST("/conversazioni/:conversazioneID/partecipanti", rt.wrapAuth(rt.addToGroup))
	rt.router.DELETE("/conversazioni/:conversazioneID/partecipanti/:partecipanteID", rt.wrapAuth(rt.leaveGroup))

	// ---------------- MESSAGGI E COMMENTI ----------------
	rt.router.POST("/conversazioni/:conversazioneID/messaggi", rt.wrapAuth(rt.sendMessage))

	rt.router.PATCH("/conversazioni/:conversazioneID/status", rt.wrapAuth(rt.updateChatStatus))

	// RISOLUZIONE CONFLITTO: L'URL del forward è stato modificato per non collidere con :messaggioID
	rt.router.POST("/conversazioni/:conversazioneID/forward", rt.wrapAuth(rt.forwardMessage))

	rt.router.DELETE("/conversazioni/:conversazioneID/messaggi/:messaggioID", rt.wrapAuth(rt.deleteMessage))

	// Commenti/Reazioni
	rt.router.POST("/conversazioni/:conversazioneID/messaggi/:messaggioID/commenti", rt.wrapAuth(rt.commentMessage))
	rt.router.DELETE("/conversazioni/:conversazioneID/messaggi/:messaggioID/commenti/:commentoID", rt.wrapAuth(rt.uncommentMessage))
}

// ServeHTTP makes the router implement the http.Handler interface.
func (rt *_router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.router.ServeHTTP(w, r)
}
