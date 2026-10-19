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
	Owner       string    `json:"owner,omitempty"`   // Username del creatore/proprietario del gruppo
	Members     []string  `json:"members,omitempty"` // Lista degli username dei partecipanti
	LastMessage string    `json:"lastMessage,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	UnreadCount int       `json:"unreadCount"` // Aggiunto per le notifiche stile WhatsApp
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
	Status     string    `json:"status,omitempty"` // Aggiunto per le doppie spunte ('sent', 'received', 'read')
}

// Comment rappresenta una reazione a un messaggio.
type Comment struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}
