package store

import (
	"context"
	"os"
	"strconv"
	"time"

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
}
