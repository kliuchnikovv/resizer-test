package main

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type cachingStatus string

const (
	pending  cachingStatus = "pending"
	finished cachingStatus = "finished"
	failed   cachingStatus = "failed"
)

type cacheItem struct {
	status cachingStatus

	data []byte
}

type workerRequest struct {
	url    string
	width  uint
	height uint
	key    string
}

type service struct {
	cache *lru.Cache

	numberOfWorkers int

	retrieveAttempts int
	retrieveTimeoutS time.Duration

	wg     *sync.WaitGroup
	cancel context.CancelFunc

	workChan chan workerRequest
}

type resizeRequest struct {
	URLs   []string `json:"urls"`
	Width  uint     `json:"width"`
	Height uint     `json:"height"`
}

type resizeResult struct {
	Result string `json:"result"`
	URL    string `json:"url,omitempty"`
	Cached bool   `json:"cached"`
}
