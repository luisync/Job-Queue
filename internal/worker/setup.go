package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Create the redis consumer group.
func CreateConsumerGroup(ctx context.Context, redisClient *redis.Client, streamKey string, groupName string) error {
	if err := redisClient.XGroupCreateMkStream(ctx, streamKey, groupName, "0").Err(); err != nil {
		// Group already exists.
		if strings.Contains(err.Error(), "BUSYGROUP") {
			slog.Debug("Consumer group already exists", "group", groupName, "straem", streamKey)
			return nil
		}
		return fmt.Errorf("Failed to create consumer group %s on %s, %w", groupName, streamKey, err)
	}

	slog.Info("Conusmer group created", "group", groupName, "stream", streamKey)
	return nil
}
