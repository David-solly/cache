package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type MemoryCache struct {
	client *Store
}

type Store struct {
	name  string
	data  map[string]interface{}
	mutex sync.Mutex
}

func (c MemoryCache) NewClient() *Store {
	return &Store{name: "Memory store",
		data: map[string]interface{}{"PING": "PONG"}}
}

func (c *MemoryCache) init() (string, error) {
	c.client = c.NewClient()
	pong, _ := c.client.data["PING"]

	fmt.Println("Memory store - Online ..........")
	return pong.(string), nil
}

func (c *MemoryCache) Initialise() (string, error) {
	return c.init()
}

func (c *MemoryCache) StoreRecord(model Record) (bool, error) {
	c.client.mutex.Lock()
	if model.ValueMap != nil {
		c.client.data[strings.ToUpper(model.Key)] = model.ValueMap
	} else {
		c.client.data[strings.ToUpper(model.Key)] = strings.ToUpper(model.Value)
	}
	c.client.mutex.Unlock()
	return true, nil
}

func (c *MemoryCache) ReadCache(key string) (interface{}, bool, error) {
	c.client.mutex.Lock()
	data, k := c.client.data[strings.ToUpper(key)]
	c.client.mutex.Unlock()
	if !k {
		return "", false, fmt.Errorf("Value @ key: '%q' - Not Found", key)
	}
	return data, true, nil
}

// StoreExpiringRecord :
// Creates a sleeping gorouting that will awake and delete
// stored value found with 'k' only after 'duration'
func (c *MemoryCache) StoreExpiringRecord(model Expirer) (bool, error) {
	k, v, t := model.GetExpiringRecord()
	c.client.mutex.Lock()
	c.client.data[strings.ToUpper(k)] = v
	c.client.mutex.Unlock()
	go func(k string, duration time.Duration, c *MemoryCache) {
		time.Sleep(duration)
		c.DeleteFromCache(k)
		return
	}(k, t, c)
	return true, nil
}

func (c *MemoryCache) DeleteFromCache(key string) (bool, error) {
	c.client.mutex.Lock()
	delete(c.client.data, strings.ToUpper(key))
	c.client.mutex.Unlock()
	return true, nil
}
