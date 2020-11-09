package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	resp, err := http.Get("https://api.stackexchange.com/2.2/questions?pagesize=100&order=desc&sort=activity&tagged=go&site=stackoverflow&filter=withbody")
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
	someClcn := pGodb.Collection("stackArticles")

	for i, v := range list {
		user := v["owner"].(map[string]interface{})

		body := v["body"].(string)
		para := body[3:strings.Index(body, "</p>")]

		doc := bson.D{
			{"title", v["title"]},
			{"subtitle", fmt.Sprintf("User: %v", user["display_name"])},
			{"content", fmt.Sprintf("{%d} %s\nlink: %v", i, para, v["link"])},
			{"timestamp", time.Now()},
		}

		_, err := someClcn.InsertOne(ctx, doc)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Inserted %v into collection.\n", i)
	}
}
