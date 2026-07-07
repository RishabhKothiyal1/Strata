package schema

import (
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type PubSub struct {
	nc *nats.Conn
}

func NewPubSub(natsURL string) (*PubSub, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	slog.Info("Successfully connected to NATS JetStream server for GraphQL subscriptions.")
	return &PubSub{nc: nc}, nil
}

func (p *PubSub) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}

func (p *PubSub) Publish(subject string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to serialize NATS publish message", "subject", subject, "error", err)
		return
	}

	if err := p.nc.Publish(subject, payload); err != nil {
		slog.Error("Failed to publish message to NATS", "subject", subject, "error", err)
	}
}

// Subscribe returns a channel of raw map payloads and a unsubscribe/cleanup function.
func (p *PubSub) Subscribe(subject string) (<-chan map[string]interface{}, func(), error) {
	out := make(chan map[string]interface{}, 10)

	sub, err := p.nc.Subscribe(subject, func(m *nats.Msg) {
		var msg map[string]interface{}
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			slog.Error("Failed to unmarshal NATS subscription payload", "subject", subject, "error", err)
			return
		}
		out <- msg
	})
	if err != nil {
		close(out)
		return nil, nil, err
	}

	cleanup := func() {
		sub.Unsubscribe()
		close(out)
		slog.Debug("Unsubscribed from NATS topic", "subject", subject)
	}

	return out, cleanup, nil
}
