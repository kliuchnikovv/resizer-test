package main

import (
	"context"
	"fmt"
	"log"
	"sync"
)

func (s *service) startWorkers(ctx context.Context) error {
	if s.numberOfWorkers <= 0 {
		return fmt.Errorf("invalid number of workers: %d", s.numberOfWorkers)
	}

	s.wg = &sync.WaitGroup{}
	ctx, s.cancel = context.WithCancel(ctx)
	s.workChan = make(chan workerRequest, 2*s.numberOfWorkers)

	for i := 0; i < s.numberOfWorkers; i++ {
		s.wg.Add(1)
		go worker(ctx, s.wg, s.workChan, s.cache.Add)
	}

	return nil
}

func (s *service) stopWorkers() error {
	if s.cancel == nil {
		return fmt.Errorf("already cancelled")
	}

	s.cancel()
	s.wg.Wait()

	close(s.workChan)
	s.cancel = nil

	return nil
}

func worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	input <-chan workerRequest,
	addToCache func(key, value interface{}) bool,
) {
	var shouldWork bool = true

	log.Print("worker started")

	for shouldWork {
		select {
		case request, ok := <-input:
			if !ok {
				log.Print("input channel closed")
				shouldWork = false

				break
			}

			if err := fetchAndResize(request.url, request.width, request.height, request.key, addToCache); err != nil {
				log.Printf("failed to resize %s: %v", request.url, err)
			}

			log.Printf("processed request: %s", request.url)
		case <-ctx.Done():
			shouldWork = false
		}
	}

	wg.Done()
	log.Print("worker stopped")
}
