package db

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"strconv"
	"time"
	"log"
)

const newestJobKey = "_newest_job"

type r struct {
	c *redis.Client
}

func New(host string) (*r, error) {
	options, err := redis.ParseURL(host)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	err = client.Ping().Err()
	if err != nil {
		return nil, err
	}

	return &r{c: client}, nil
}

func (r *r) Lock(key string, jobId int, expiration time.Duration, ctx context.Context) error {
	locked, err := r.tryLock(key, jobId, expiration)
	if err != nil {
		return err
	}
	if locked {
		return nil
	}

	log.Printf("somebody else has locked the resource, so trying to get lock for %v", expiration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			locked, err := r.tryLock(key, jobId, expiration)
			if err != nil {
				return err
			}
			if locked {
				return nil
			}
			break
		}
	}
}

func (r *r) tryLock(key string, jobId int, expiration time.Duration) (bool, error) {
	var newestJob int
	var lockId int
	getLockedId, err := r.c.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			newestJob = 0
		}
	} else {
		lockId, err = strconv.Atoi(getLockedId)
		if err != nil {
			return false, err
		}
		if lockId == jobId {
			return false, errors.Errorf("Job %v seems to be executed twice", lockId)
		}
	}

	getNewestJobResult, err := r.c.Get(key + newestJobKey).Result()
	if err != nil {
		if err == redis.Nil {
			newestJob = 0
		} else {
			return false, err
		}
	}

	if getNewestJobResult == "" {
		newestJob = 0
	} else {
		newestJob, err = strconv.Atoi(getNewestJobResult)
		if err != nil {
			return false, err
		}
	}

	if newestJob > jobId {
		return false, errors.Errorf("Newer Job: %v > Actual Job: %v is also waiting for a lock.", newestJob, jobId)
	} else if newestJob < jobId {
		r.c.Set(key+newestJobKey, jobId, 0).Err()
	}

	set, err := r.c.SetNX(key, jobId, expiration).Result()
	if err != nil {
		return false, err
	}

	if set {
		log.Print("we aquired the lock")
		return true, nil
	}

	return false, nil
}
