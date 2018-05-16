package db

type Client interface {
	Lock() error
	Unlock() error
}
