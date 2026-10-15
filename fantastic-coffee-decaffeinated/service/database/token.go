package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// CheckToken verifica se il token (ID utente) esiste nel database.
func (db *appdbimpl) CheckToken(token string) (bool, error) {
	var id string
	err := db.c.QueryRow("SELECT id FROM users WHERE id = ?", token).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Il token non è associato a nessun utente
			return false, nil
		}
		// C'è stato un problema col DB
		return false, fmt.Errorf("error checking token: %w", err)
	}
	// Token trovato!
	return true, nil
}
