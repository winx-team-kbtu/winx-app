package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.Equal(t, 0, len(hub.clients))
}

func TestHub_Register(t *testing.T) {
	hub := NewHub()
	client := &Client{
		userID: 1,
		send:   make(chan []byte, 256),
		conn:   nil,
		hub:    hub,
	}

	hub.Register(client)
	assert.Equal(t, 1, len(hub.clients))
	assert.Equal(t, client, hub.clients[1])
}

func TestHub_Register_ReplacesOldConnection(t *testing.T) {
	hub := NewHub()

	oldClient := &Client{
		userID: 1,
		send:   make(chan []byte, 256),
		conn:   nil,
		hub:    hub,
	}

	newClient := &Client{
		userID: 1,
		send:   make(chan []byte, 256),
		conn:   nil,
		hub:    hub,
	}

	hub.Register(oldClient)
	assert.Equal(t, 1, len(hub.clients))
	assert.Equal(t, oldClient, hub.clients[1])

	hub.Register(newClient)
	assert.Equal(t, 1, len(hub.clients))
	assert.Equal(t, newClient, hub.clients[1])
}

func TestHub_Unregister(t *testing.T) {
	hub := NewHub()
	client := &Client{
		userID: 1,
		send:   make(chan []byte, 256),
		conn:   nil,
		hub:    hub,
	}

	hub.Register(client)
	assert.Equal(t, 1, len(hub.clients))

	hub.Unregister(1)
	assert.Equal(t, 0, len(hub.clients))
}

func TestHub_Send(t *testing.T) {
	hub := NewHub()
	client := &Client{
		userID: 1,
		send:   make(chan []byte, 256),
		conn:   nil,
		hub:    hub,
	}
	hub.Register(client)

	payload := []byte(`{"type":"test"}`)
	hub.Send(1, payload)

	select {
	case msg := <-client.send:
		assert.Equal(t, payload, msg)
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout: Expected message not received")
	}
}

func TestHub_Send_NonExistentUser(t *testing.T) {
	hub := NewHub()
	hub.Send(999, []byte(`{"type":"test"}`))
}

func TestHub_Send_BufferFull(t *testing.T) {
	hub := NewHub()
	client := &Client{
		userID: 1,
		send:   make(chan []byte, 1),
		conn:   nil,
		hub:    hub,
	}
	hub.Register(client)

	client.send <- []byte(`{"type":"first"}`)

	hub.Send(1, []byte(`{"type":"second"}`))

	select {
	case msg := <-client.send:
		assert.Equal(t, []byte(`{"type":"first"}`), msg)
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout: Expected first message not received")
	}

	select {
	case <-client.send:
	default:
	}
}

func TestClient_SendError(t *testing.T) {
	hub := NewHub()
	client := &Client{
		userID: 1,
		send:   make(chan []byte, 256),
		conn:   nil,
		hub:    hub,
	}
	hub.Register(client)

	client.sendError("test error")

	select {
	case msg := <-client.send:
		assert.Contains(t, string(msg), "error")
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout: Expected error message not received")
	}
}
