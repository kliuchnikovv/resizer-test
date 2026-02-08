package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"time"

	jpgresize "github.com/nfnt/resize"
)

func (s *service) processResizes(request resizeRequest, isAsync bool) ([]resizeResult, error) {
	results := make([]resizeResult, 0, len(request.URLs))
	for _, url := range request.URLs {
		result, err := s.processResize(url, request.Width, request.Height, isAsync)
		if err != nil {
			log.Printf("failed to resize %s: %v", url, err)

			result = &resizeResult{Result: failure}
		}

		results = append(results, *result)
	}

	return results, nil
}

func (s *service) processResize(
	url string,
	width uint,
	height uint,
	isAsync bool,
) (*resizeResult, error) {
	var (
		id     = genID(url, width, height)
		key    = "/v1/image/" + id + ".jpeg"
		newURL = proto + hostport + key
	)

	if s.cache.Contains(key) {
		log.Printf("found in cache %s", url)

		return &resizeResult{
			URL:    newURL,
			Result: success,
			Cached: true,
		}, nil
	}

	if isAsync {
		select {
		case s.workChan <- workerRequest{
			url:    url,
			width:  width,
			height: height,
			key:    key,
		}:
		default:
			return nil, fmt.Errorf("queue is full")
		}
	} else {
		if err := fetchAndResize(url, width, height, key, s.cache.Add); err != nil {
			return nil, err
		}
	}

	return &resizeResult{
		URL:    newURL,
		Result: success,
	}, nil
}

func (s *service) processRetrieve(url string) ([]byte, error) {
	log.Print("fetching ", url)

	for i := 0; i < s.retrieveAttempts; i++ {
		result, err := s.getCacheItem(url)
		if err == nil && result.status == finished {
			log.Printf("found in cache (%d): %v", i, err)
			return result.data, nil
		}

		if err != nil {
			log.Printf("not found in cache (%d)", i)
		} else {
			log.Printf("wrong status: %s", result.status)
		}

		// sleep for timeout
		time.Sleep(s.retrieveTimeoutS * time.Second)
	}

	log.Printf("exceeded attempts")
	return nil, fmt.Errorf("not found")
}

func (s *service) getCacheItem(url string) (*cacheItem, error) {
	data, ok := s.cache.Get(url)
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	result, ok := data.(cacheItem)
	if !ok {
		return nil, fmt.Errorf("invalid type")
	}

	return &result, nil
}

func fetchAndResize(
	url string,
	width uint,
	height uint,
	key string,
	addToCache func(key, value interface{}) bool,
) error {
	addToCache(key, cacheItem{
		status: pending,
	})

	data, err := fetch(url)
	if err != nil {
		addToCache(key, cacheItem{
			status: failed,
		})
		log.Printf("failed to fetch %s: %v", url, err)

		return err
	}

	resizedData, err := resize(data, width, height)
	if err != nil {
		addToCache(key, cacheItem{
			status: failed,
		})
		log.Printf("failed to resize %s: %v", url, err)

		return err
	}

	log.Print("caching ", key)
	addToCache(key, cacheItem{
		status: finished,
		data:   resizedData,
	})

	return nil
}

func fetch(url string) ([]byte, error) {
	log.Print("fetching ", url)
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status: %d", r.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, 15*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read fetch data: %v", err)
	}

	return data, nil
}

func resize(data []byte, width uint, height uint) ([]byte, error) {
	// decode jpeg into image.Image
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to jped decode: %v", err)
	}

	// if either width or height is 0, it will resize respecting the aspect ratio
	var newImage = jpgresize.Resize(width, height, img, jpgresize.Lanczos3)

	newData := bytes.Buffer{}
	err = jpeg.Encode(bufio.NewWriter(&newData), newImage, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to jpeg encode resized image: %v", err)
	}

	return newData.Bytes(), nil
}

func genID(url string, width uint, height uint) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d-%d", url, width, height)))
	return base64.URLEncoding.EncodeToString(hash[:])
}
