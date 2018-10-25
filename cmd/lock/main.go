package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cosee-gitlab/lock"
	"github.com/cosee-gitlab/lock/db"
	log "github.com/sirupsen/logrus"
	"time"
)

var lockKey = flag.String("key", "", "(optional) locking key")
var doLock = flag.Bool("lock", false, "should I aquire a lock? Returns 0 if lock is aquired and not 0 if error occured or lock wasn't granted")
var doUnlock = flag.Bool("unlock", false, "unlock the locked thing")
var lockExpiration = flag.Duration("expiration", 15*time.Minute, "time after the lock gets removed, even if unlock isn't called")
var scope = flag.String("scope", "job", "can be job, stage or project")
var verbose = flag.Int("v", 3, "set verbosity level (3: Warning, 4: Info, 5: Debug)")

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	flag.Parse()

	fmt.Printf("cosee-gitlab/lock: %v, commit %v, built at %v", version, commit, date)

	if *verbose < 0 || 5 < *verbose {
		log.Fatalf("Illegal verbosity level: %v", *verbose)
		return
	}
	log.SetLevel(log.Level(*verbose))

	env, err := lock.LoadEnvironment()
	if err != nil {
		log.Fatal(err)
		return
	}

	if env.JobId == 0 || env.PipelineId == 0 || env.ProjectId == 0 || env.JobName == "" {
		log.Fatal("It seems that no GitLab environment variables are set")
		return
	}

	if env.RedisHost == "" {
		log.Fatal("Please set LOCKS_REDIS_HOST variable")
		return
	}

	if *doLock == *doUnlock {
		log.Fatal("I can't lock and unlock . Doin' nothing isn't an option, too")
		return
	}

	client, err := db.New(env.RedisHost)
	if err != nil {
		log.Fatal(err)
		return
	}

	var key string
	if len(*lockKey) != 0 {
		key = *lockKey
	} else {
		switch *scope {
		case "job":
			key = fmt.Sprintf("%d-%s", env.ProjectId, env.JobName)
		case "stage":
			key = fmt.Sprintf("%d-s:%s", env.ProjectId, env.JobStage)
		case "project":
			key = fmt.Sprintf("%d", env.ProjectId)
		default:
			log.Fatalf("%q isn't a valid scope. Scope can be job, stage or project.", *scope)
			return
		}
	}

	if *doLock {
		err = client.Lock(key, env.JobId, *lockExpiration, context.Background())
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	if *doUnlock {
		err = client.Unlock(key, env.JobId)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
