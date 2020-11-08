package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	resp, err := http.Get("https://api.stackexchange.com/2.2/questions?pagesize=100&order=desc&sort=activity&tagged=go&site=stackoverflow")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result map[string][]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	list := result["items"]

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("TEST_ATLAS_URI")))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	pGodb := client.Database("pointyGo")
	someClcn := pGodb.Collection("articles")

	for i, v := range list {
		doc := bson.D{
			{"title", v["title"]},
			{"subtitle", fmt.Sprintf("Views: %.0f", v["view_count"])},
			{"content", fmt.Sprintf("Lorem ipsum content here.  %d", i)},
			{"timestamp", time.Now()},
		}

		_, err := someClcn.InsertOne(ctx, doc)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Inserted %v into collection.\n", i)
	}
}
