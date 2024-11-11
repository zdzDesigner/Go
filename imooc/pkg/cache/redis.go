package cache

import (
	"github.com/go-redis/redis"
)

type SubCB func(string, string)

type Cacher interface {
	Pub(key, value string)
	Sub(key string, cb SubCB)
}

type cache struct {
	client *redis.Client
}

func (c *cache) Pub(key, value string) {
	c.client.Publish(key, value).Err()
}

func (c *cache) Sub(key string, fn SubCB) {
	pubsub := c.client.Subscribe(key)
	_, err := pubsub.Receive()
	if err != nil {
		return
	}
	go func() {
		for msg := range pubsub.Channel() {
			fn(msg.Channel, msg.Payload)
		}
	}()
}

func NewCache() Cacher {
	return &cache{client: Client()}
}

var redisClient *redis.Client

func Client() *redis.Client {
	if redisClient == nil {
		redisClient = newRedis()
	}
	return redisClient
}

func newRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   13,
	})

	if err := redisClient.Ping().Err(); err != nil {
		panic(err)
	}
	return redisClient
}
