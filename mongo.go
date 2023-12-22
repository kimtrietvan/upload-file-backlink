package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectToServer() *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error %s\n", err)
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://"+user+":"+pass+"@"+host+":"+port))
	if err != nil {
		panic(err)
	}
	connection := client.Database(os.Getenv("DB_DATABASE")).Collection("template")
	return connection
}
