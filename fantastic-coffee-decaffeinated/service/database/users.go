package database

import (
	"errors"
	"fmt"
	"strings"
)

// SetMyUserName aggiorna lo username dell'utente. Fallisce se il nome esiste già.
func (db *appdbimpl) SetMyUserName(userID string, newName string) error {
	// Tentiamo l'aggiornamento. Se "newName" è già presente in un altro record,
	// il constraint UNIQUE nella tabella users farà fallire la query.
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", newName, userID)
	if err != nil {
		// Controlliamo se l'errore è dovuto alla violazione del vincolo UNIQUE di SQLite
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("conflict: username already in use")
		}
		return fmt.Errorf("error updating username: %w", err)
	}

	return nil
}

// SetMyPhoto aggiorna la foto profilo dell'utente.
func (db *appdbimpl) SetMyPhoto(userID string, photo string) error {
	_, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", photo, userID)
	if err != nil {
		return fmt.Errorf("error updating user photo: %w", err)
	}
	return nil
}
