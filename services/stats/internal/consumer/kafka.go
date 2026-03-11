package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/repository"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	repo   repository.EventRepository
}

func NewKafkaConsumer(addr, topic, groupID string, repo repository.EventRepository) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{addr},
			Topic:   topic,
			GroupID: groupID,
		}),
		repo: repo,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("Failed to read message", "error", err)
			continue
		}

		var event domain.ClickEvent
		if err = json.Unmarshal(msg.Value, &event); err != nil {
			slog.Error("Failed to unmarshal event", "error", err)
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if event.ClickedAt.IsZero() {
			event.ClickedAt = time.Now().UTC()
		}

		if err = c.repo.Save(ctx, &event); err != nil {
			slog.Error("Failed to save event", "error", err)
			select {
			case <-time.After(5 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		err = c.reader.CommitMessages(ctx, msg)
		if err != nil {
			slog.Error("Failed to commit message", "error", err)
			continue
		}

		slog.Info("Click event saved", "code", event.Code, "country", event.Country)
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
