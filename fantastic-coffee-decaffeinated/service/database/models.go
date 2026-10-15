package database

import "time"

// User rappresenta un utente del sistema WASAText.
type User struct {
	ID    string `json:"identifier,omitempty"` // Usato per il Bearer token
	Name  string `json:"name"`
	Photo string `json:"photo,omitempty"`
}

// Conversation rappresenta una chat (singola o di gruppo) con i suoi metadati.
type Conversation struct {
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	Photo       string    `json:"photo,omitempty"`
	LastMessage string    `json:"lastMessage,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
}

// Message rappresenta un singolo messaggio all'interno di una conversazione.
type Message struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Sender     string    `json:"sender"`
	Content    string    `json:"content"`
	Type       string    `json:"type"` // "text" o "photo"
	Checkmarks int       `json:"checkmarks"`
	Reactions  []Comment `json:"reactions,omitempty"`
}

// Comment rappresenta una reazione a un messaggio.
type Comment struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}
