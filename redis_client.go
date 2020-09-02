package cache

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-redis/redis"
)

type RedisCache struct {
	client *redis.Client
}

// Init : redis
func (c *RedisCache) init() (string, error) {
	// From Deployment or environmental variables
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		return "", errors.New("No address supplied")
	}
	c.client = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	k, err := c.client.Ping().Result()
	if err != nil {
		return k, err
	}

	fmt.Println("Redis server - Online ..........")
	return k, nil
}

func (c *RedisCache) Initialise() (string, error) {
	return c.init()

}

func (c *RedisCache) StoreRecord(model Record) (bool, error) {
	if c.client == nil {
		return false, errors.New("Redis client is nil")
	}
	base := c.client.Set(strings.ToUpper(model.Key), strings.ToUpper(model.Value), 0)
	errAccess := base.Err()
	if errAccess != nil {
		return false, errAccess
	}
	return true, nil
}

// StoreExpiringRecord :
// Creates a sleeping gorouting that will awake and delete
// stored value found with 'k' only after 'duration'
func (c *RedisCache) StoreExpiringRecord(model Expirer) (bool, error) {
	k, v, t := model.GetExpiringRecord()

	base := c.client.Set(strings.ToUpper(k), v, t)
	errAccess := base.Err()
	if errAccess != nil {
		return false, errAccess
	}
	return true, nil
}

func (c *RedisCache) ReadCache(key string) (interface{}, bool, error) {
	data, err := c.client.Get(strings.ToUpper(key)).Result()

	if err != nil {
		return "", false, fmt.Errorf("Value @ key: '%q' - Not Found", key)
	}
	return data, true, nil
}

func (c *RedisCache) DeleteFromCache(key string) (bool, error) {
	base := c.client.Del(strings.ToUpper(key))
	err := base.Err()
	if err != nil {
		return false, fmt.Errorf("Error deleteing value at @ key: '%q'", key)
	}

	return true, nil
}
