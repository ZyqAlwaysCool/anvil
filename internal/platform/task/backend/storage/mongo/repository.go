package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/zyq/anvil/internal/platform/task"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(ctx context.Context, collection *mongo.Collection) (*Repository, error) {
	repo := &Repository{collection: collection}
	if err := repo.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *Repository) Create(ctx context.Context, record task.Record) error {
	if _, err := r.collection.InsertOne(ctx, record); err != nil {
		return fmt.Errorf("insert task record: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, taskID string) (task.Record, bool, error) {
	var record task.Record
	err := r.collection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&record)
	if err == nil {
		return record, true, nil
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		return task.Record{}, false, nil
	}
	return task.Record{}, false, fmt.Errorf("find task record: %w", err)
}

func (r *Repository) SetQueued(ctx context.Context, taskID, queueMessageID string) error {
	return r.update(ctx, taskID, bson.M{
		"$set": bson.M{
			"status":           task.StatusQueued,
			"queue_message_id": queueMessageID,
			"updated_at":       time.Now().UTC(),
		},
	})
}

func (r *Repository) SetRunning(ctx context.Context, taskID string, startedAt time.Time) error {
	return r.update(ctx, taskID, bson.M{
		"$set": bson.M{
			"status":     task.StatusRunning,
			"started_at": startedAt,
			"updated_at": startedAt,
		},
	})
}

func (r *Repository) SetSuccess(ctx context.Context, taskID string, result json.RawMessage, finishedAt time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"status":      task.StatusSuccess,
			"updated_at":  finishedAt,
			"finished_at": finishedAt,
		},
	}
	if len(result) > 0 {
		update["$set"].(bson.M)["result"] = result
	}
	return r.update(ctx, taskID, update)
}

func (r *Repository) SetFailed(ctx context.Context, taskID, errMsg string, finishedAt time.Time) error {
	return r.update(ctx, taskID, bson.M{
		"$set": bson.M{
			"status":        task.StatusFailed,
			"error_message": errMsg,
			"updated_at":    finishedAt,
			"finished_at":   finishedAt,
		},
	})
}

func (r *Repository) MergeMetadata(ctx context.Context, taskID string, fields map[string]any) error {
	setDoc := bson.M{"updated_at": time.Now().UTC()}
	for key, value := range fields {
		setDoc["metadata."+key] = value
	}
	return r.update(ctx, taskID, bson.M{"$set": setDoc})
}

func (r *Repository) IncrementTokenUsage(ctx context.Context, taskID string, inputDelta, outputDelta int) error {
	incDoc := bson.M{}
	if inputDelta > 0 {
		incDoc[task.MetadataTokenInputKey] = inputDelta
	}
	if outputDelta > 0 {
		incDoc[task.MetadataTokenOutputKey] = outputDelta
	}
	if len(incDoc) == 0 {
		return nil
	}
	return r.update(ctx, taskID, bson.M{
		"$inc": incDoc,
		"$set": bson.M{"updated_at": time.Now().UTC()},
	})
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("count task records: %w", err)
	}
	return count, nil
}

func (r *Repository) update(ctx context.Context, taskID string, update bson.M) error {
	result, err := r.collection.UpdateByID(ctx, taskID, update)
	if err != nil {
		return fmt.Errorf("update task record: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("task record not found: %s", taskID)
	}
	return nil
}

func (r *Repository) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}}, Options: options.Index().SetName("idx_status")},
		{Keys: bson.D{{Key: "agent", Value: 1}}, Options: options.Index().SetName("idx_agent")},
		{Keys: bson.D{{Key: "type", Value: 1}}, Options: options.Index().SetName("idx_type")},
		{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("idx_created_at")},
	}
	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create task indexes: %w", err)
	}
	return nil
}

var _ task.Repository = (*Repository)(nil)
