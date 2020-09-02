package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/api/option"

	storage "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

type FirestoreCache struct {
	client         *storage.Client
	CollectionName string
	ValueKey       string
}

// Init : redis
func (c *FirestoreCache) init() (string, error) {
	var app *firebase.App
	var err error
	ctx := context.Background()

	// From Deployment or environmental variables
	projectID := os.Getenv("FIRESTORE_PROJECT_ID")
	if len(projectID) == 0 {
		dsn := option.WithCredentialsFile("./.ac-plugin-k.json")
		app, err = firebase.NewApp(ctx, nil, dsn)
	} else {
		conf := &firebase.Config{ProjectID: projectID}
		app, err = firebase.NewApp(ctx, conf)
	}
	if err != nil {
		log.Fatalln(err)
	}

	c.client, err = app.Firestore(ctx)
	if err != nil {
		if c.client != nil {
			c.client.Close()
		}
		c.client = nil
		return "", err
	}

	c.CollectionName = "license-keys"
	c.ValueKey = "value"

	fmt.Println("Firestore - Online ..........")
	return "PONG", nil
}

func (c *FirestoreCache) Initialise() (string, error) {
	return c.init()

}

func (c *FirestoreCache) StoreRecord(model Record) (bool, error) {
	if c.client == nil {
		return false, errors.New("Firebase client is nil")
	}
	ctx := context.Background()
	_, err := c.client.Collection(c.CollectionName).Doc(strings.ToUpper(model.Key)).Set(ctx, map[string]interface{}{
		"value": strings.ToUpper(model.Value),
	})
	if err != nil {
		return false, fmt.Errorf("Failed adding record:%q with error: %v", model.Key, err)
	}

	return true, nil
}

// StoreExpiringRecord :
// Creates a sleeping gorouting that will awake and delete
// // stored value found with 'k' only after 'duration'
func (c *FirestoreCache) StoreExpiringRecord(model Expirer) (bool, error) {
	// k, v, t := model.GetExpiringRecord()

	// base := c.client.Set(strings.ToUpper(k), v, t)
	// errAccess := base.Err()
	// if errAccess != nil {
	// 	return false, errAccess
	// }
	return false, nil
}

func (c *FirestoreCache) ReadCache(key string) (string, bool, error) {
	data, err := c.client.Collection(c.CollectionName).Doc(strings.ToUpper(key)).Get(context.Background())
	if err != nil {
		return "", false, fmt.Errorf("Value @ key: '%q' - Not Found", key)
	}
	m := data.Data()
	fmt.Printf("Document data: %#v\n", m)
	return m[c.ValueKey].(string), data.Exists(), nil
}
