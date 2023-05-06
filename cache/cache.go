package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/sync/singleflight"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache *redis.Client
	// use singleflight.Group to make sure that each key is only fetched once.
	loader *singleflight.Group
}

var (
	// Allow multiple readers or a single writer.
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: RedisClient,
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {

	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	if v, err := g.mainCache.Get(key).Bytes(); err == nil {
		log.Println("[Cache] hit")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) HGet(key, field string) (b []byte, err error) {
	if key == "" || field == "" {
		return nil, fmt.Errorf("key or field is required")
	}
	if v, err := g.mainCache.HGet(key, field).Bytes(); err == nil {
		log.Println("[Cache] hit")
		return v, nil
	}

	// Situation: cache not exist
	return g.load(key + field)
}

func (g *Group) load(key string) (b []byte, err error) {

	viewi, err, _ := g.loader.Do(key, func() (interface{}, error) {
		// TODO: distributed nodes support
		return g.getFromDatabase(key)
	})

	if err != nil {
		return nil, err

	}
	return viewi.([]byte), nil

}

func (g *Group) getFromDatabase(key string) ([]byte, error) {
	value, err := g.getter.Get(key)
	if err != nil {
		return nil, err
	}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value []byte) {
	g.mainCache.Set(key, value, time.Hour)
}
