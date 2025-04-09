package bots

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Gotify struct {
	url      string
	token    string
	title    string
	priority int
	client   *http.Client
}

type gotifyMessage struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// NewGotify creates a new Gotify bot instance
func NewGotify(url, token, title string, priority int) (*Gotify, error) {
	if url == "" {
		return nil, fmt.Errorf("gotify URL cannot be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("gotify token cannot be empty")
	}

	// Default title if not provided
	if title == "" {
		title = "ED-AFK-Notifier"
	}

	// Default priority if not specified or invalid
	if priority <= 0 {
		priority = 5 // Medium priority as default
	}

	return &Gotify{
		url:      url,
		token:    token,
		title:    title,
		priority: priority,
		client:   &http.Client{},
	}, nil
}

// Start satisfies the Bot interface but does nothing for Gotify
// as it doesn't need to listen for incoming messages
func (g *Gotify) Start() {
	log.Info("Gotify notification service ready")
}

// Send sends a message to Gotify server
func (g *Gotify) Send(text string) error {
	return g.SendWithPriority(text, g.priority)
}

// SendWithPriority sends a message to Gotify server with a specific priority
func (g *Gotify) SendWithPriority(text string, priority int) error {
	message := gotifyMessage{
		Title:    g.title,
		Message:  text,
		Priority: priority,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	url := fmt.Sprintf("%s/message?token=%s", g.url, g.token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
