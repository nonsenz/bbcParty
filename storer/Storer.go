package storer

import (
	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"time"
)

type Storer interface {
	Get(key string, bucket string) string
	Put(key string, value string, bucket string) error
	All(bucket string) []string
	Random(bucket string) string
	CreateBucket(bucket string) error
	Close()
}

type BoltStorer struct {
	db bolt.DB
}

func NewBoltStorer(dbName string) *BoltStorer {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &BoltStorer{*db}
}

func (b *BoltStorer) Close() {
	b.db.Close()
}

func (b *BoltStorer) Get(key string, bucket string) string {
	var value string
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		value = string(b.Get([]byte(key)))
		return nil
	})

	return value
}

func (b *BoltStorer) Put(key string, value string, bucket string) error {
	if err := b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			log.Fatal(err)
			return err
		}
		bucket := tx.Bucket([]byte(bucket))
		if err := bucket.Put([]byte(key), []byte(value)); err != nil {
			log.Fatal(err)
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (b *BoltStorer) All(bucket string) []string {
	var keys []string
	if err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if(bucket != nil) {
			// Iterate over items in sorted key order.
			if err := bucket.ForEach(func(key, value []byte) error {
				keys = append(keys, string(key))
				return nil
			}); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	return keys
}

func (b *BoltStorer) Random(bucket string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	values := b.All(bucket)
	if (len(values) == 0) {
		return ""
	}
	return values[rand.Intn(len(values))]
}

func (b *BoltStorer) CreateBucket(bucket string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	});

	return err
}
