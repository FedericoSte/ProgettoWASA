package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		// Consenti le richieste dal frontend Vue (in sviluppo)
		handlers.AllowedOrigins([]string{"http://localhost:5173", "http://127.0.0.1:5173"}),

		// Consenti i metodi HTTP necessari per la tua API
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT", "PATCH"}),

		// FONDAMENTALE: Consenti l'header "Authorization" per far passare il Bearer Token
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(h)
}
