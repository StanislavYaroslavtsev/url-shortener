package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/services/url-shortener/internal/geo"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
	geo    *geo.IP
}

func NewKafkaProducer(addr, topic string, geo *geo.IP) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(addr),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 5 * time.Second,
		},
		geo: geo,
	}
}

func (p *KafkaProducer) SendClickEvent(ctx context.Context, link *domain.Link, ip string) error {
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	country := p.geo.Country(ip)
	event := map[string]any{
		"code":       link.Code,
		"ip":         ip,
		"country":    country,
		"clicked_at": time.Now().UTC(),
	}

	slog.Info("Sending click event", "code", link.Code, "ip", ip, "country", country)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	}); err != nil {
		return fmt.Errorf("failed to send click event: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
