package cache

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/docker/docker/pkg/testutil/assert"
)

// set to a reachable redis instance if one is available
// otherwise ignored
const globalRedis = "192.168.99.100:6379"
const useRedis = false

func TestInitialiseCache(t *testing.T) {
	c := Cache{}
	t.Run("INITIALISE cache", func(t *testing.T) {
		suite := []struct {
			testName  string
			addr      string
			use, want bool
			cacheType Service
			err       string
		}{
			{"INITIALISE - redis", globalRedis, true, true, &RedisCache{}, ""},
			{"INITIALISE - redis", "192.168.99.100:6349", true, false, &RedisCache{}, "No connection"},
			{"INITIALISE - redis", globalRedis, false, true, &MemoryCache{}, ""},
			{"INITIALISE - firestore", "", true, true, &FirestoreCache{}, ""},
			{"INITIALISE - memory", "", false, true, &MemoryCache{}, ""},
		}
		for i, test := range suite {
			if !useRedis && test.testName == "INITIALISE - redis" {
				fmt.Println("Skipping redis check")
				continue
			}
			t.Run(fmt.Sprintf("#%d: %q%q", i, test.testName, test.addr), func(t *testing.T) {
				ok, err := c.Initialise(test.addr, test.use)
				if test.err == "" {
					assert.NotNil(t, c.Client)
					assert.NilError(t, err)
					assert.Equal(t, ok, test.want)
					if !test.use {
						assert.Equal(t, reflect.TypeOf(c.Client), reflect.TypeOf(test.cacheType))
					}
					if test.use {
						assert.Equal(t, reflect.TypeOf(c.Client), reflect.TypeOf(test.cacheType))
					}

				} else {
					assert.Error(t, err, test.err)
					assert.Equal(t, ok, test.want)
				}
			})
		}
		t.Run("INITIALISE cache FAULT - redis", func(t *testing.T) {
			os.Setenv("REDIS_DSN", "")
			r := RedisCache{}
			s, e := r.Initialise()
			assert.Equal(t, s, "")
			assert.Error(t, e, "No address supplied")
		})

	})
}

func TestStoreGenerateResposne(t *testing.T) {
	c := Cache{}
	c.Initialise("", false)
	longDuration := time.Duration(time.Second * 10)
	shortDuration := time.Duration(time.Millisecond * 300)
	suite := []struct {
		testName string
		data     RecordExpirer
		expect   bool
		err      string
	}{
		{"CACHE -  ", RecordExpirer{Key: "FFA45722AA7", Value: "38240", Timeout: longDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "38245", Timeout: shortDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "38245", Timeout: shortDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A12", Value: "38245", Timeout: shortDuration}, true, ""},
	}

	for i, test := range suite {
		// Retrieve full value via 5 char shortcode
		t.Run(fmt.Sprintf("#%d - SAVE CACHE: %q", i, test.data.Key), func(t *testing.T) {
			k, err := c.Client.StoreExpiringRecord(&test.data)
			assert.NilError(t, err)
			assert.DeepEqual(t, k, true)
			// device, found, err := c.Client.StoreExpiringRecord(test.data.ShortCode)
		})
	}

	suiteRead := []struct {
		testName string
		data     RecordExpirer
		expect   bool
		err      string
	}{
		{"CACHE -  ", RecordExpirer{Key: "FFA45722AA7", Value: "38240", Timeout: longDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "38245", Timeout: shortDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A111", Value: "", Timeout: shortDuration}, false, "Not Found"},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A12", Value: "38245", Timeout: shortDuration}, false, "Not Found"},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "", Timeout: shortDuration}, false, "Not Found"},
	}
	for i, test := range suiteRead {

		time.Sleep(100 * time.Millisecond)

		t.Run(fmt.Sprintf("#%d - READ CACHE: %q", i, test.data.Key), func(t *testing.T) {
			s, k, err := c.Client.ReadCache(test.data.Key)
			assert.DeepEqual(t, k, test.expect)
			if test.err != "" {
				assert.Error(t, err, test.err)
			} else {
				assert.NilError(t, err)
			}

			if k {
				assert.DeepEqual(t, s, test.data.Value)
			}

		})
	}
}

func TestStoreGenerateResposneREDIS(t *testing.T) {
	c := Cache{}
	c.Initialise(globalRedis, useRedis)
	longDuration := time.Duration(time.Second * 10)
	shortDuration := time.Duration(time.Millisecond * 300)
	suite := []struct {
		testName string
		data     RecordExpirer
		expect   bool
		err      string
	}{
		{"CACHE -  ", RecordExpirer{Key: "FFA45722AA7", Value: "38240", Timeout: longDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "38245", Timeout: shortDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "38245", Timeout: shortDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A12", Value: "38245", Timeout: shortDuration}, true, ""},
	}

	for i, test := range suite {
		// Retrieve full value via 5 char shortcode
		t.Run(fmt.Sprintf("#%d - SAVE CACHE: %q", i, test.data.Key), func(t *testing.T) {
			k, err := c.Client.StoreExpiringRecord(&test.data)
			assert.NilError(t, err)
			assert.DeepEqual(t, k, true)

		})
	}

	suiteRead := []struct {
		testName string
		data     RecordExpirer
		expect   bool
		err      string
	}{
		{"CACHE -  ", RecordExpirer{Key: "FFA45722AA7", Value: "38240", Timeout: longDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "38245", Timeout: shortDuration}, true, ""},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A111", Value: "", Timeout: shortDuration}, false, "Not Found"},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A12", Value: "38245", Timeout: shortDuration}, false, "Not Found"},
		{"CACHE -  ", RecordExpirer{Key: "FFA45722A11", Value: "", Timeout: shortDuration}, false, "Not Found"},
	}
	for i, test := range suiteRead {

		time.Sleep(100 * time.Millisecond)

		t.Run(fmt.Sprintf("#%d - READ CACHE: %q", i, test.data.Key), func(t *testing.T) {
			s, k, err := c.Client.ReadCache(test.data.Key)
			assert.DeepEqual(t, k, test.expect)
			if test.err != "" {
				assert.Error(t, err, test.err)
			} else {
				assert.NilError(t, err)
			}

			if k {
				assert.DeepEqual(t, s, test.data.Value)
			}

		})
	}
}

func TestCache5chars(t *testing.T) {
	mCache := Cache{}
	mCache.Initialise("", false)

	fCache := Cache{}
	fCache.Initialise("", true)

	rCache := Cache{}
	// Redis endpoint - true flag to confirm redis as choice
	rCache.Initialise(globalRedis, useRedis)

	t.Run("SAVE and READ from CACHE", func(t *testing.T) {
		suite := []struct {
			testName string
			data     Record
			expect   bool
			cache    Cache
			err      string
		}{
			{"CACHE - memory ", Record{Value: "FFA45722AA7", Key: "38240"}, true, mCache, ""},
			{"CACHE - redis", Record{Value: "FFA45722AA7", Key: "38241"}, true, rCache, ""},
			{"CACHE - redis", Record{Value: "FFA45722AA7", Key: "38251"}, false, Cache{Client: &RedisCache{}}, ""},
			{"CACHE - memory", Record{Value: "FFA45722AA7", Key: "38242"}, true, mCache, ""},
			{"CACHE - memory", Record{Value: "FFA45722AA7", Key: "38212"}, true, mCache, ""},
			{"CACHE - firestore", Record{Value: "FFA45722AA7", Key: "38212"}, true, fCache, ""},
		}

		for i, test := range suite {
			t.Run(fmt.Sprintf("#%d - %q: %q", i, test.testName, test.data.Key), func(t *testing.T) {

				ok, err := test.cache.Client.StoreRecord(test.data)
				assert.Equal(t, ok, test.expect)
				if !test.expect {
					assert.Error(t, err, test.err)
				} else {
					assert.NilError(t, err)
				}
			})
		}

		t.Run("READ from CACHE", func(t *testing.T) {
			suite := []struct {
				testName string
				data     Record
				expect   bool
				cache    Cache
				err      string
			}{
				{"CACHE - memory ", Record{Value: "FFA45722AA7", Key: "38240"}, true, mCache, ""},
				{"CACHE - redis", Record{Value: "FFA45722AA7", Key: "38241"}, true, rCache, ""},
				{"CACHE - memory", Record{Value: "FFA45722AA7", Key: "38225"}, false, mCache, "Value @ key: '\"38225\"' - Not Found"},
				{"CACHE - memory", Record{Value: "FFA45722AA7", Key: "38225"}, false, rCache, "Value @ key: '\"38225\"' - Not Found"},
				{"CACHE - firestore", Record{ValueMap: map[string]interface{}{"value": "FFA45722AA7"}, Key: "38212"}, true, fCache, ""},
			}

			for i, test := range suite {
				// Retrieve full value via 5 char shortcode
				t.Run(fmt.Sprintf("#%d - READ CACHE: %q", i, test.data.Key), func(t *testing.T) {
					device, found, err := test.cache.Client.ReadCache(test.data.Key)
					assert.Equal(t, found, test.expect)
					if test.testName == "CACHE - firestore" {
						assert.Equal(t, reflect.TypeOf(device), reflect.TypeOf(map[string]interface{}{}))
					} else {
						assert.Equal(t, reflect.TypeOf(device), reflect.TypeOf("deveui"))
					}
					if test.expect {
						assert.NilError(t, err)
						if test.testName == "CACHE - firestore" {
							assert.DeepEqual(t, device, test.data.ValueMap)
						} else {
							assert.Equal(t, device, test.data.Value)
						}
					}
					if !test.expect {
						assert.Error(t, err, test.err)
						if test.testName == "CACHE - firestore" {
							assert.Equal(t, device, nil)
						} else {
							assert.Equal(t, device, "")
						}

					}

				})
			}

		})
	})
}
