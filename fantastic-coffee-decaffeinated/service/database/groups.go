package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AddToGroup aggiunge un utente a un gruppo esistente.
// Controlla prima che il chiamante (callerID) faccia già parte del gruppo.
func (db *appdbimpl) AddToGroup(conversationID string, userToAdd string, callerID string) error {
	// 1. Controllo sicurezza: il chiamante deve essere già nel gruppo per poter aggiungere altri
	var exists int
	err := db.c.QueryRow(`
		SELECT 1 FROM participants 
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, callerID).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("unauthorized: caller is not in the group")
		}
		return fmt.Errorf("error checking caller permissions: %w", err)
	}

	// 2. Controllo validità: l'utente da aggiungere esiste?
	var userExists int
	err = db.c.QueryRow("SELECT 1 FROM users WHERE name = ?", userToAdd).Scan(&userExists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("not found: user to add does not exist")
		}
		return fmt.Errorf("error checking if user exists: %w", err)
	}

	// 3. Aggiungiamo l'utente al gruppo
	// Usiamo INSERT OR IGNORE per evitare errori se l'utente è già nel gruppo
	_, err = db.c.Exec(`
		INSERT OR IGNORE INTO participants (conversation_id, user_id) 
		VALUES (?, (SELECT id FROM users WHERE name = ?))
	`, conversationID, userToAdd)

	if err != nil {
		return fmt.Errorf("error adding user to group: %w", err)
	}

	return nil
}

// LeaveGroup rimuove un utente (tipicamente chi effettua la richiesta) da un gruppo.
func (db *appdbimpl) LeaveGroup(conversationID string, userID string) error {
	// Elimina il record dalla tabella participants
	res, err := db.c.Exec(`
		DELETE FROM participants 
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, userID)

	if err != nil {
		return fmt.Errorf("error leaving group: %w", err)
	}

	// Controlliamo se è stato effettivamente eliminato qualcosa
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("not found: user was not in the group or group does not exist")
	}

	return nil
}

// SetGroupName aggiorna il nome di un gruppo.
func (db *appdbimpl) SetGroupName(conversationID string, newName string, callerID string) error {
	// Controllo sicurezza: il chiamante deve essere nel gruppo
	var exists int
	err := db.c.QueryRow("SELECT 1 FROM participants WHERE conversation_id = ? AND user_id = ?", conversationID, callerID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("unauthorized: caller is not in the group")
		}
		return fmt.Errorf("error checking permissions: %w", err)
	}

	_, err = db.c.Exec("UPDATE conversations SET name = ? WHERE id = ? AND type = 'group'", newName, conversationID)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}
	return nil
}

// SetGroupPhoto aggiorna la foto di un gruppo.
func (db *appdbimpl) SetGroupPhoto(conversationID string, photo string, callerID string) error {
	// Controllo sicurezza: il chiamante deve essere nel gruppo
	var exists int
	err := db.c.QueryRow("SELECT 1 FROM participants WHERE conversation_id = ? AND user_id = ?", conversationID, callerID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("unauthorized: caller is not in the group")
		}
		return fmt.Errorf("error checking permissions: %w", err)
	}

	_, err = db.c.Exec("UPDATE conversations SET photo = ? WHERE id = ? AND type = 'group'", photo, conversationID)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}
	return nil
}
