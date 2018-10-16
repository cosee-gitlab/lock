package db

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
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
	log := log.WithField("key", key)
	log.Debug("-lock flag was set, so trying to acquire lock")

	locked, err := r.tryLock(key, jobId, expiration)
	if err != nil {
		return err
	}
	if locked {
		return nil
	}

	log.Warnf("somebody else has already locked the resource, so trying to get lock for %v", expiration)

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

func (r *r) Unlock(key string, jobId int) error {
	log := log.WithField("key", key)
	log.Debug("-unlock flag was set, so trying to unlock")

	var lockId int
	getLockId, err := r.c.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			lockId = 0
		}
	} else {
		lockId, err = strconv.Atoi(getLockId)
		if err != nil {
			return err
		}
		if lockId != jobId {
			return errors.Errorf("Job %v has the lock. Thats not me.", lockId)
		}
	}

	delRes, err := r.c.Del(key).Result()
	if err != nil {
		return err
	}

	if delRes != 1 {
		log.Warn("Couldn't remove lock. It isn't there")
	}

	return nil
}

func (r *r) tryLock(key string, jobId int, expiration time.Duration) (bool, error) {
	log := log.WithField("key", key)
	log.Info("Trying...")

	var newestJob int
	var lockId int
	getLockId, err := r.c.Get(key).Result()
	log.WithField("getLockId", getLockId).WithError(err).Debugf("Get(key:%v)", key)
	if err != nil {
		if err == redis.Nil {
			lockId = 0
		}
	} else {
		lockId, err = strconv.Atoi(getLockId)
		if err != nil {
			return false, err
		}
		if lockId == jobId {
			return false, errors.Errorf("Job %v seems to be executed twice", lockId)
		}
	}

	getNewestJobResult, err := r.c.Get(key + newestJobKey).Result()
	log.WithField("getNewestJobResult", getNewestJobResult).WithError(err).Debugf("Get(key:%v)", key+newestJobKey)
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
		err = r.c.Set(key+newestJobKey, jobId, 0).Err()
		log.WithError(err).Debugf("Set(key:%v, jobId:%v, 0)", key, jobId)
	}

	set, err := r.c.SetNX(key, jobId, expiration).Result()
	log.WithError(err).WithField("logAcquired", set).Debugf("SetNX(key:%v, jobId:%v, expiration:%v)", key, jobId, expiration)
	if err != nil {
		return false, err
	}

	if set {
		log.Print("we aquired the lock")
		return true, nil
	}

	return false, nil
}
