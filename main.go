package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title,omitempty"`
	Subtitle  string             `bson:"subtitle,omitempty"`
	Content   string             `bson:"content,omitempty"`
	Timestamp time.Time          `bson:"timestamp,omitempty"`
}

var client *mongo.Client
var articlesCollection *mongo.Collection
var err error

type Router struct {
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("TEST_ATLAS_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	articlesCollection = client.Database("pointyGo").Collection("articles")

	mux := &Router{}
	http.ListenAndServe(":9090", mux)
}

func (p *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	case "/articles":
		fetchArticles(w, r)
	default:
		http.NotFound(w, r)
	}
	return
}

func fetchArticles(resp http.ResponseWriter, req *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	resp.Header().Add("content-type", "application/json")

	switch method := req.Method; method {
	case "GET":
		filterCursor, err := articlesCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Print(err)
		}
		defer filterCursor.Close(ctx)

		var articles []Article
		if err = filterCursor.All(ctx, &articles); err != nil {
			log.Print(err)
		}

		json.NewEncoder(resp).Encode(articles)

	case "POST":
		var article Article
		json.NewDecoder(req.Body).Decode(&article)
		article.Timestamp = time.Now()
		respEncode := json.NewEncoder(resp)
		if result, err := articlesCollection.InsertOne(ctx, article); err != nil {
			log.Print(err)
			respEncode.Encode("{status: error}")
		} else {
			respEncode.Encode(result)
		}
	}
}

// fmt.Println("Found,")
// for filterCursor.Next(ctx) {
// 	var article Article
// 	if err := filterCursor.Decode(&article); err != nil {
// 		panic(err)
// 	}
// 	fmt.Print(article.Title, "\n")
// }
