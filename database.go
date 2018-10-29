package main

import (
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
	"log"
)

var db *mongo.Database

func connect() (bool, error) {
	con, err := mongo.NewClient("mongodb://admin:HardAdminPassword1@ds129386.mlab.com:29386/paragliding-cloud2")

	con.Connect(context.Background())
	if err != nil {
		return false, err;
	}
	db = con.Database("paragliding-cluster")
	return true, err;
}

func insert(track Track) interface{} {
	collection := db.Collection("tracks")

	res, err := collection.InsertOne(context.Background(), &track)
	if err != nil {
		log.Fatal(err)
	}

	return res.InsertedID
}
