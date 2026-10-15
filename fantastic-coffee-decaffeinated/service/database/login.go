package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
)

// DoLogin verifica se l'utente esiste. Se sì, ritorna il suo ID.
// Altrimenti, crea un nuovo utente, genera un UUID come identificatore e lo ritorna.
func (db *appdbimpl) DoLogin(username string) (string, error) {
	var userID string

	// 1. Controlliamo se l'utente esiste già
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", username).Scan(&userID)
	if err == nil {
		// Utente trovato! Ritorna il suo ID.
		return userID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		// C'è stato un errore di esecuzione della query (diverso da "utente non trovato")
		return "", fmt.Errorf("error checking if user exists: %w", err)
	}

	// 2. L'utente non esiste (sql.ErrNoRows), dobbiamo crearlo.
	// Generiamo un nuovo UUID versione 4
	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("error generating new UUID for user: %w", err)
	}
	userID = newUUID.String()

	// Inseriamo il nuovo utente nel database
	_, err = db.c.Exec("INSERT INTO users (id, name, photo) VALUES (?, ?, '')", userID, username)
	if err != nil {
		return "", fmt.Errorf("error inserting new user: %w", err)
	}

	// Restituiamo il nuovo ID generato
	return userID, nil
}
