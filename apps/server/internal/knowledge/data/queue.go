package data

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/go-kratos/kratos/v2/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

const ingestionQueueName = "ragdesk.ingestion"
const (
	ingestionRetryHeader = "x-retry"
	defaultMaxRetries    = 3
	defaultBackoffBaseMs = 500
	envMaxRetries        = "RAGDESK_INGESTION_MAX_RETRIES"
	envBackoffBaseMs     = "RAGDESK_INGESTION_BACKOFF_MS"
)

type rabbitQueue struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
	log   *log.Helper
}

func NewIngestionQueue(cfg *conf.Data, logger log.Logger) biz.IngestionQueue {
	if cfg == nil || cfg.Rabbitmq == nil || cfg.Rabbitmq.Addr == "" {
		return nil
	}
	helper := log.NewHelper(logger)
	conn, err := amqp.Dial(cfg.Rabbitmq.Addr)
	if err != nil {
		helper.Warnf("rabbitmq dial failed: %v", err)
		return nil
	}
	ch, err := conn.Channel()
	if err != nil {
		helper.Warnf("rabbitmq channel failed: %v", err)
		_ = conn.Close()
		return nil
	}
	if _, err := ch.QueueDeclare(
		ingestionQueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		helper.Warnf("rabbitmq declare queue failed: %v", err)
		_ = ch.Close()
		_ = conn.Close()
		return nil
	}
	return &rabbitQueue{
		conn:  conn,
		ch:    ch,
		queue: ingestionQueueName,
		log:   helper,
	}
}

func (q *rabbitQueue) Enqueue(ctx context.Context, job biz.IngestionJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.publish(ctx, payload, amqp.Table{})
}

func (q *rabbitQueue) Start(ctx context.Context, handler func(context.Context, biz.IngestionJob) error) error {
	if err := q.ch.Qos(1, 0, false); err != nil {
		return err
	}
	msgs, err := q.ch.Consume(
		q.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	go func() {
		maxRetries := ingestionMaxRetries()
		backoffBase := ingestionBackoff()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var job biz.IngestionJob
				if err := json.Unmarshal(msg.Body, &job); err != nil {
					_ = msg.Nack(false, false)
					continue
				}
				if err := handler(context.Background(), job); err != nil {
					retry := getRetryCount(msg.Headers)
					if retry < maxRetries {
						delay := backoffBase
						if retry > 0 {
							delay = backoffBase * time.Duration(1<<retry)
						}
						if delay > 0 {
							time.Sleep(delay)
						}
						headers := cloneHeaders(msg.Headers)
						headers[ingestionRetryHeader] = retry + 1
						if err := q.publish(context.Background(), msg.Body, headers); err != nil {
							_ = msg.Nack(false, true)
							continue
						}
						_ = msg.Ack(false)
						continue
					}
					_ = msg.Nack(false, false)
					continue
				}
				_ = msg.Ack(false)
			}
		}
	}()
	return nil
}

func (q *rabbitQueue) publish(ctx context.Context, payload []byte, headers amqp.Table) error {
	if headers == nil {
		headers = amqp.Table{}
	}
	return q.ch.PublishWithContext(
		ctx,
		"",
		q.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         payload,
			Headers:      headers,
		},
	)
}

func ingestionMaxRetries() int {
	value := strings.TrimSpace(os.Getenv(envMaxRetries))
	if value == "" {
		return defaultMaxRetries
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return defaultMaxRetries
	}
	if n > 10 {
		return 10
	}
	return n
}

func ingestionBackoff() time.Duration {
	value := strings.TrimSpace(os.Getenv(envBackoffBaseMs))
	if value == "" {
		return time.Duration(defaultBackoffBaseMs) * time.Millisecond
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return time.Duration(defaultBackoffBaseMs) * time.Millisecond
	}
	return time.Duration(n) * time.Millisecond
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if raw, ok := headers[ingestionRetryHeader]; ok {
		switch v := raw.(type) {
		case int32:
			return int(v)
		case int64:
			return int(v)
		case int:
			return v
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return 0
}

func cloneHeaders(headers amqp.Table) amqp.Table {
	out := amqp.Table{}
	for k, v := range headers {
		out[k] = v
	}
	return out
}

func (q *rabbitQueue) Close() error {
	if q.ch != nil {
		_ = q.ch.Close()
	}
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}
