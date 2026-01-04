package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/go-kratos/kratos/v2/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

const ingestionQueueName = "ragdesk.ingestion"

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
		},
	)
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
					_ = msg.Nack(false, true)
					continue
				}
				_ = msg.Ack(false)
			}
		}
	}()
	return nil
}
