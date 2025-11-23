package jobs

import (
	"context"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func VerifyQueueConnection(config *conf.Config) error {
	conn, err := amqp.Dial(config.RabbitConn)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func AddAll(db storage.Storage, config *conf.Config) error {
	logger := logging.Get()
	conn, err := amqp.Dial(config.RabbitConn)
	if err != nil {
		logger.Errorw("Failed to connect to RabbitMQ", zap.Error(err))
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logger.Errorw("Failed to open a channel", zap.Error(err))
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"tasks", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		logger.Errorw("Failed to declare a queue", zap.Error(err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "all"
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	if err != nil {
		logger.Errorw("Failed to publish a message", zap.Error(err))
		return err
	}
	logger.Infow("Published a message", "body", body)
	return nil
}

func ReadTasks(ctx context.Context, db storage.Storage, config *conf.Config) {
	logger := logging.Get()
	conn, err := amqp.Dial(config.RabbitConn)
	if err != nil {
		logger.Errorw("Failed to connect to RabbitMQ", zap.Error(err))
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logger.Errorw("Failed to open a channel", zap.Error(err))
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"tasks", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		logger.Errorw("Failed to declare a queue", zap.Error(err))
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		logger.Errorw("Failed to register a consumer", zap.Error(err))
		return
	}

	logger.Info("Waiting for messages. Consumer started.")

	for {
		select {
		case d, ok := <-msgs:
			// Check if the channel was closed by RabbitMQ (e.g., connection lost)
			if !ok {
				logger.Error("RabbitMQ delivery channel closed unexpectedly.")
				return
			}
			msg := string(d.Body)
			logger.Infow("Received a message", "body", msg)

			if msg == "all" {
				now := time.Now()
				err = RunAllJobs(db, config, now)
				if err != nil {
					logger.Errorw("Error when running jobs.", zap.Error(err))
				}
			} else {
				logger.Errorw("Unexpected message", "body", msg)
			}

			d.Ack(false)

		case <-ctx.Done():
			// Context was cancelled (e.g., from an external signal)
			logger.Info("Context cancelled. Stopping consumer.")

			return
		}
	}
}
