//go:build ignore

package task

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"anvil-scaffold-template/internal/platform/config"
	"anvil-scaffold-template/internal/platform/logging"
)

type Worker struct {
	repo        Repository
	queue       Queue
	registry    *Registry
	logger      *slog.Logger
	consumer    string
	concurrency int
	retryJobs   bool
	maxTries    int
	jobTimeout  time.Duration
}

func NewWorker(repo Repository, queue Queue, registry *Registry, logger *slog.Logger, cfg config.TaskConfig) *Worker {
	maxTries := cfg.MaxTries
	if maxTries <= 0 {
		maxTries = 1
	}
	return &Worker{
		repo:        repo,
		queue:       queue,
		registry:    registry,
		logger:      logger,
		consumer:    cfg.ConsumerName,
		concurrency: cfg.WorkerConcurrency,
		retryJobs:   cfg.RetryJobs,
		maxTries:    maxTries,
		jobTimeout:  cfg.JobTimeout,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	var waitGroup sync.WaitGroup
	for i := 0; i < w.concurrency; i++ {
		waitGroup.Add(1)
		consumerName := fmt.Sprintf("%s-%d", w.consumer, i+1)
		go func(name string) {
			defer waitGroup.Done()
			w.consumeLoop(ctx, name)
		}(consumerName)
	}

	<-ctx.Done()
	waitGroup.Wait()
	return nil
}

func (w *Worker) consumeLoop(ctx context.Context, consumerName string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		messages, err := w.queue.Read(ctx, consumerName, 1)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("read queue failed", "consumer_name", consumerName, "error", err)
			time.Sleep(time.Second)
			continue
		}
		for _, item := range messages {
			if err := w.handleMessage(ctx, consumerName, item); err != nil {
				w.logger.Error("handle message failed", "task_id", item.Message.TaskID, "error", err)
			}
		}
	}
}

// handleMessage 固定流转：SetRunning -> Handle -> Retry/SetFailed -> Ack。
func (w *Worker) handleMessage(ctx context.Context, consumerName string, item QueuedMessage) error {
	taskLogger := w.logger.With("task_id", item.Message.TaskID, "type", item.Message.Type, "consumer_name", consumerName)
	taskCtx := logging.WithLogger(ctx, taskLogger)

	startedAt := time.Now().UTC()
	if err := w.repo.SetRunning(taskCtx, item.Message.TaskID, startedAt); err != nil {
		return err
	}

	handler, ok := w.registry.Lookup(item.Message.Type)
	if !ok {
		finishedAt := time.Now().UTC()
		if err := w.repo.SetFailed(taskCtx, item.Message.TaskID, "task handler not registered", finishedAt); err != nil {
			return err
		}
		return w.queue.Ack(taskCtx, item.QueueMessageID)
	}

	handleCtx := taskCtx
	if w.jobTimeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(taskCtx, w.jobTimeout)
		defer cancel()
		handleCtx = timeoutCtx
	}

	result, err := handler.Handle(handleCtx, item.Message)
	if err != nil {
		attempt := item.Message.Attempt
		if attempt <= 0 {
			attempt = 1
		}
		if w.retryJobs && attempt < w.maxTries {
			retryMessage := item.Message
			retryMessage.Attempt = attempt + 1
			newQueueMessageID, retryErr := w.queue.Retry(taskCtx, item.QueueMessageID, retryMessage)
			if retryErr != nil {
				if newQueueMessageID != "" {
					// 部分成功：重试消息已入队，但旧 delivery 未 Ack；旧消息可能被再次消费，handler 必须幂等。
					taskLogger.Warn("task retry partial success",
						"new_queue_message_id", newQueueMessageID,
						"old_queue_message_id", item.QueueMessageID,
						"error", retryErr,
					)
					return retryErr
				}
				finishedAt := time.Now().UTC()
				if updateErr := w.repo.SetFailed(taskCtx, item.Message.TaskID, retryErr.Error(), finishedAt); updateErr != nil {
					return updateErr
				}
				return retryErr
			}
			// Retry 成功：新消息已入队且旧 delivery 已 Ack，结束当前消费轮次。
			return nil
		}

		finishedAt := time.Now().UTC()
		if updateErr := w.repo.SetFailed(taskCtx, item.Message.TaskID, err.Error(), finishedAt); updateErr != nil {
			return updateErr
		}
		return w.queue.Ack(taskCtx, item.QueueMessageID)
	}

	finishedAt := time.Now().UTC()
	if err := w.repo.SetSuccess(taskCtx, item.Message.TaskID, result, finishedAt); err != nil {
		return err
	}
	return w.queue.Ack(taskCtx, item.QueueMessageID)
}
