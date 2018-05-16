package lock

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Environment struct {
	ProjectId   int   `envconfig:"CI_PROJECT_ID"`
	ProjectName string `envconfig:"CI_PROJECT_NAME"`
	PipelineId  int   `envconfig:"CI_PIPELINE_ID"`
	JobId       int   `envconfig:"CI_JOB_ID"`
	JobName     string `envconfig:"CI_JOB_NAME"`

	RedisHost string `envconfig:"LOCKS_REDIS_HOST"`
}

func LoadEnvironment() (*Environment, error) {
	var e Environment
	err := envconfig.Process("lock", &e)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load gitlab environment variables")
	}

	return &e, nil
}
