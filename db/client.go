package db

import (
	"time"
	"context"
)

type Client interface {
	Lock(key string, jobId int, expiration time.Duration, ctx context.Context) error
	Unlock(key string, jobId int) error
}
