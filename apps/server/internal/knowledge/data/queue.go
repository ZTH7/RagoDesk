package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/go-kratos/kratos/v2/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

const (
	ingestionQueueName      = "ragdesk.ingestion"
	ingestionRetryQueueName = "ragdesk.ingestion.retry"
	ingestionDLQName        = "ragdesk.ingestion.dlq"
)
const (
	ingestionRetryHeader = "x-retry"
	defaultMaxRetries    = 3
	defaultBackoffBaseMs = 500
	defaultWorkerCount   = 1
	envMaxRetries        = "RAGDESK_INGESTION_MAX_RETRIES"
	envBackoffBaseMs     = "RAGDESK_INGESTION_BACKOFF_MS"
	envWorkerCount       = "RAGDESK_INGESTION_WORKERS"
)

type rabbitQueue struct {
	conn        *amqp.Connection
	ch          *amqp.Channel
	queue       string
	retry       string
	dlq         string
	log         *log.Helper
	maxRetries  int
	backoffBase time.Duration
	workerCount int
	pubMu       sync.Mutex
}

func NewIngestionQueue(cfg *conf.Data, logger log.Logger) biz.IngestionQueue {
	if cfg == nil {
		return nil
	}
	maxRetries, backoffBase, workerCount := resolveIngestionRuntimeConfig(cfg)
	if cfg.Rabbitmq != nil && cfg.Rabbitmq.Addr != "" {
		helper := log.NewHelper(logger)
		conn, err := amqp.Dial(cfg.Rabbitmq.Addr)
		if err != nil {
			helper.Warnf("rabbitmq dial failed: %v", err)
		} else {
			ch, err := conn.Channel()
			if err != nil {
				helper.Warnf("rabbitmq channel failed: %v", err)
				_ = conn.Close()
			} else {
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
				} else if _, err := ch.QueueDeclare(
					ingestionRetryQueueName,
					true,
					false,
					false,
					false,
					amqp.Table{
						"x-dead-letter-exchange":    "",
						"x-dead-letter-routing-key": ingestionQueueName,
					},
				); err != nil {
					helper.Warnf("rabbitmq declare retry queue failed: %v", err)
					_ = ch.Close()
					_ = conn.Close()
				} else if _, err := ch.QueueDeclare(
					ingestionDLQName,
					true,
					false,
					false,
					false,
					nil,
				); err != nil {
					helper.Warnf("rabbitmq declare dlq failed: %v", err)
					_ = ch.Close()
					_ = conn.Close()
				} else {
					return &rabbitQueue{
						conn:        conn,
						ch:          ch,
						queue:       ingestionQueueName,
						retry:       ingestionRetryQueueName,
						dlq:         ingestionDLQName,
						log:         helper,
						maxRetries:  maxRetries,
						backoffBase: backoffBase,
						workerCount: workerCount,
					}
				}
			}
		}
	}
	return newRedisQueue(cfg, logger, maxRetries, backoffBase, workerCount)
}

func (q *rabbitQueue) Enqueue(ctx context.Context, job biz.IngestionJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.publish(ctx, q.queue, payload, amqp.Table{}, 0)
}

func (q *rabbitQueue) Start(ctx context.Context, handler func(context.Context, biz.IngestionJob) error) error {
	if q.conn == nil {
		return errors.New("rabbitmq connection missing")
	}
	workers := q.workerCount
	if workers <= 0 {
		workers = defaultWorkerCount
	}
	for i := 0; i < workers; i++ {
		go q.consume(ctx, handler)
	}
	return nil
}

func (q *rabbitQueue) consume(ctx context.Context, handler func(context.Context, biz.IngestionJob) error) {
	ch, err := q.conn.Channel()
	if err != nil {
		q.log.Warnf("rabbitmq worker channel failed: %v", err)
		return
	}
	defer func() { _ = ch.Close() }()
	if err := ch.Qos(1, 0, false); err != nil {
		q.log.Warnf("rabbitmq worker qos failed: %v", err)
		return
	}
	msgs, err := ch.Consume(
		q.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		q.log.Warnf("rabbitmq consume failed: %v", err)
		return
	}
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
			if err := handler(ctx, job); err != nil {
				retry := getRetryCount(msg.Headers)
				if retry < q.maxRetries {
					delay := q.backoffBase
					if retry > 0 {
						delay = q.backoffBase * time.Duration(1<<retry)
					}
					headers := cloneHeaders(msg.Headers)
					headers[ingestionRetryHeader] = retry + 1
					if err := q.publish(context.Background(), q.retry, msg.Body, headers, delay); err != nil {
						_ = msg.Nack(false, true)
						continue
					}
					_ = msg.Ack(false)
					continue
				}
				headers := cloneHeaders(msg.Headers)
				headers[ingestionRetryHeader] = retry
				if err := q.publish(context.Background(), q.dlq, msg.Body, headers, 0); err != nil {
					_ = msg.Nack(false, true)
					continue
				}
				_ = msg.Ack(false)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func (q *rabbitQueue) publish(ctx context.Context, queue string, payload []byte, headers amqp.Table, delay time.Duration) error {
	if q.ch == nil {
		return errors.New("rabbitmq channel missing")
	}
	if headers == nil {
		headers = amqp.Table{}
	}
	pub := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         payload,
		Headers:      headers,
	}
	if delay > 0 {
		pub.Expiration = fmt.Sprintf("%d", delay.Milliseconds())
	}
	q.pubMu.Lock()
	defer q.pubMu.Unlock()
	return q.ch.PublishWithContext(
		ctx,
		"",
		queue,
		false,
		false,
		pub,
	)
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

func (q *rabbitQueue) Health(ctx context.Context) error {
	if q.conn == nil || q.conn.IsClosed() {
		return errors.New("rabbitmq connection closed")
	}
	return nil
}

type redisQueue struct {
	client      *redis.Client
	queue       string
	dlq         string
	log         *log.Helper
	maxRetries  int
	backoffBase time.Duration
	workerCount int
}

type redisIngestionPayload struct {
	Job   biz.IngestionJob `json:"job"`
	Retry int              `json:"retry,omitempty"`
}

func newRedisQueue(cfg *conf.Data, logger log.Logger, maxRetries int, backoffBase time.Duration, workerCount int) biz.IngestionQueue {
	if cfg == nil || cfg.Redis == nil || cfg.Redis.Addr == "" {
		return nil
	}
	helper := log.NewHelper(logger)
	options := &redis.Options{
		Addr: cfg.Redis.Addr,
	}
	if cfg.Redis.Network != "" {
		options.Network = cfg.Redis.Network
	}
	if cfg.Redis.ReadTimeout != nil {
		options.ReadTimeout = cfg.Redis.ReadTimeout.AsDuration()
	}
	if cfg.Redis.WriteTimeout != nil {
		options.WriteTimeout = cfg.Redis.WriteTimeout.AsDuration()
	}
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		helper.Warnf("redis ping failed: %v", err)
		_ = client.Close()
		return nil
	}
	return &redisQueue{
		client:      client,
		queue:       ingestionQueueName,
		dlq:         ingestionDLQName,
		log:         helper,
		maxRetries:  maxRetries,
		backoffBase: backoffBase,
		workerCount: workerCount,
	}
}

func (q *redisQueue) Enqueue(ctx context.Context, job biz.IngestionJob) error {
	payload, err := json.Marshal(redisIngestionPayload{Job: job})
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, q.queue, payload).Err()
}

func (q *redisQueue) Start(ctx context.Context, handler func(context.Context, biz.IngestionJob) error) error {
	workers := q.workerCount
	if workers <= 0 {
		workers = defaultWorkerCount
	}
	for i := 0; i < workers; i++ {
		go q.consume(ctx, handler)
	}
	return nil
}

func (q *redisQueue) consume(ctx context.Context, handler func(context.Context, biz.IngestionJob) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		result, err := q.client.BRPop(ctx, time.Second, q.queue).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, redis.Nil) {
				continue
			}
			q.log.Warnf("redis brpop failed: %v", err)
			continue
		}
		if len(result) < 2 {
			continue
		}
		job, retry, err := decodeRedisPayload(result[1])
		if err != nil {
			q.log.Warnf("redis payload decode failed: %v", err)
			continue
		}
		if err := handler(ctx, job); err != nil {
			if retry < q.maxRetries {
				delay := q.backoffBase
				if retry > 0 {
					delay = q.backoffBase * time.Duration(1<<retry)
				}
				q.requeueWithDelay(job, retry+1, delay)
			} else {
				q.pushDLQ(job, retry)
			}
		}
	}
}

func (q *redisQueue) requeueWithDelay(job biz.IngestionJob, retry int, delay time.Duration) {
	payload, err := json.Marshal(redisIngestionPayload{Job: job, Retry: retry})
	if err != nil {
		q.log.Warnf("redis requeue marshal failed: %v", err)
		return
	}
	if delay <= 0 {
		if err := q.client.RPush(context.Background(), q.queue, payload).Err(); err != nil {
			q.log.Warnf("redis requeue failed: %v", err)
		}
		return
	}
	time.AfterFunc(delay, func() {
		if err := q.client.RPush(context.Background(), q.queue, payload).Err(); err != nil {
			q.log.Warnf("redis delayed requeue failed: %v", err)
		}
	})
}

func (q *redisQueue) pushDLQ(job biz.IngestionJob, retry int) {
	payload, err := json.Marshal(redisIngestionPayload{Job: job, Retry: retry})
	if err != nil {
		q.log.Warnf("redis dlq marshal failed: %v", err)
		return
	}
	if err := q.client.RPush(context.Background(), q.dlq, payload).Err(); err != nil {
		q.log.Warnf("redis dlq push failed: %v", err)
	}
}

func decodeRedisPayload(raw string) (biz.IngestionJob, int, error) {
	var payload redisIngestionPayload
	if err := json.Unmarshal([]byte(raw), &payload); err == nil {
		if payload.Job.DocumentID != "" || payload.Job.DocumentVersionID != "" {
			return payload.Job, payload.Retry, nil
		}
	}
	var job biz.IngestionJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return biz.IngestionJob{}, 0, err
	}
	return job, 0, nil
}

func (q *redisQueue) Close() error {
	if q.client == nil {
		return nil
	}
	return q.client.Close()
}

func (q *redisQueue) Health(ctx context.Context) error {
	if q.client == nil {
		return errors.New("redis client missing")
	}
	return q.client.Ping(ctx).Err()
}

func resolveIngestionRuntimeConfig(cfg *conf.Data) (int, time.Duration, int) {
	maxRetries := defaultMaxRetries
	backoff := time.Duration(defaultBackoffBaseMs) * time.Millisecond
	workerCount := defaultWorkerCount
	if cfg != nil && cfg.Knowledge != nil && cfg.Knowledge.Ingestion != nil {
		ingestion := cfg.Knowledge.Ingestion
		if ingestion.MaxRetries > 0 {
			maxRetries = int(ingestion.MaxRetries)
		}
		if ingestion.BackoffBaseMs > 0 {
			backoff = time.Duration(ingestion.BackoffBaseMs) * time.Millisecond
		}
		if ingestion.WorkerConcurrency > 0 {
			workerCount = int(ingestion.WorkerConcurrency)
		}
	}
	if value := strings.TrimSpace(os.Getenv(envMaxRetries)); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			maxRetries = n
		}
	}
	if value := strings.TrimSpace(os.Getenv(envBackoffBaseMs)); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			backoff = time.Duration(n) * time.Millisecond
		}
	}
	if value := strings.TrimSpace(os.Getenv(envWorkerCount)); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			workerCount = n
		}
	}
	if maxRetries < 0 {
		maxRetries = defaultMaxRetries
	}
	if maxRetries > 10 {
		maxRetries = 10
	}
	if backoff < 0 {
		backoff = time.Duration(defaultBackoffBaseMs) * time.Millisecond
	}
	if workerCount <= 0 {
		workerCount = defaultWorkerCount
	}
	if workerCount > 32 {
		workerCount = 32
	}
	return maxRetries, backoff, workerCount
}
