# glock
some jobs in a CI/CD pipeline just aren't designed to run in parallel

```bash
$ glock -help

Usage of glock:
  -expiration duration
        time after the lock gets removed, even if unlock isn't called (default 15m0s)
  -key string
        (optional) locking key
  -lock
        should I aquire a lock? Returns 0 if lock is aquired and not 0 if error occured or lock wasn't granted
  -scope string
        can be job, stage or project (default "job")
  -unlock
        unlock the locked thing

```