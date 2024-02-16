package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	coll *mongo.Collection
	client *mongo.Client
}

var db = &DB{}

func Init() {
	uri := "mongodb://root:example@localhost:27017"

	mps := os.Getenv("MAX_POOL_SIZE")
	maxPool, err := strconv.Atoi(mps)
	if err != nil {
		panic(err)
	}

	minPool, err := strconv.Atoi(os.Getenv("MIN_POOL_SIZE"))
	if err != nil {
		panic(err)
	}

	opts := options.Client().SetTimeout(time.Duration(time.Second * 6)).SetMaxPoolSize(uint64(maxPool)).SetMinPoolSize(uint64(minPool)).ApplyURI(uri)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	coll := client.Database("rinha").Collection("accounts")
	db.coll = coll
	db.client = client

	isSeeder := os.Getenv("IS_SEEDER")

	if isSeeder == "true" {
		err = wipeCollection()
		if err != nil {
			panic(err)
		}

		err = seedDb()
		if err != nil {
			panic(err)
		}

	}
}

func wipeCollection() error {
	fmt.Println("wiping collection")
	_, err := db.coll.DeleteMany(context.TODO(), bson.D{})
	return err
}

func seedDb() error {
	docsPath, _ := filepath.Abs("/app/seed.json")

	byteValues, err := os.ReadFile(docsPath)

	if err != nil {
		return err
	}

	var docs []interface{}

	if err = json.Unmarshal(byteValues, &docs); err != nil {
		return err
	}


	fmt.Println("seeding collection")
	_, err = db.coll.InsertMany(context.TODO(), docs)

	if err != nil {
		return err
	}

	return nil
}
