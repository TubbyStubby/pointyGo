package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type Article struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title,omitempty"`
	Subtitle  string             `bson:"subtitle,omitempty"`
	Content   string             `bson:"content,omitempty"`
	Timestamp time.Time          `bson:"timestamp,omitempty"`
}

var lock sync.Mutex
var client *mongo.Client
var articlesCollection *mongo.Collection
var err error
var index mongo.IndexModel

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

	index = mongo.IndexModel{
		Options: options.Index().SetBackground(true),
		Keys: bsonx.MDoc{
			"title":    bsonx.String("text"),
			"content":  bsonx.String("text"),
			"subtitle": bsonx.String("text"),
		},
	}
	if _, err := articlesCollection.Indexes().CreateOne(ctx, index); err != nil {
		log.Print(err)
	}

	mux := &Router{}
	http.ListenAndServe(":9090", mux)
}

func checkString(s string, rs string) bool {
	return regexp.MustCompile(rs).MatchString(s)
}

func (p *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	switch {
	case r.Method == "GET" && checkString(r.URL.Path, `^/articles`):
		fetchArticles(w, r)

	case r.Method == "POST" && checkString(r.URL.Path, `^/articles$`):
		postArticles(w, r)

	case r.Method == "GET" && checkString(r.URL.Path, `/articles/search\?.*`):
		query := r.Form["q"]
		searchArticles(w, r, query)

	case r.Method == "GET" && checkString(r.URL.Path, `^/articles/[a-zA-z0-9]*$`):
		id := r.URL.Path[len("/articles/"):]
		fetchArticles(w, r, id)

	default:
		http.NotFound(w, r)
	}
	return
}

func searchArticles(resp http.ResponseWriter, req *http.Request, query []string) {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	resp.Header().Add("content-type", "application/json")

	indexCursor, err := articlesCollection.Find(ctx, bson.M{
		"$text": bson.M{
			"$search": strings.Join(query, " "),
		},
	})
	if err != nil {
		log.Print(err)
	}

	var articles []Article

	if err = indexCursor.All(ctx, &articles); err != nil {
		log.Print(err)
	}

	json.NewEncoder(resp).Encode(articles)
}

func fetchArticles(resp http.ResponseWriter, req *http.Request, sid ...string) {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	resp.Header().Add("content-type", "application/json")

	var filterCursor *mongo.Cursor
	var err error
	if len(sid) > 0 {
		id, _ := primitive.ObjectIDFromHex(sid[0])
		filterCursor, err = articlesCollection.Find(ctx, bson.M{"_id": id})
	} else {
		findOpt := options.Find()

		if req.Form["offset"] != nil {
			skip, _ := strconv.Atoi(req.Form["offset"][0])
			findOpt.SetSkip(int64(skip))
		}
		if req.Form["limit"] != nil {
			limit, _ := strconv.Atoi(req.Form["limit"][0])
			findOpt.SetLimit(int64(limit))
		}
		filterCursor, err = articlesCollection.Find(ctx, bson.M{}, findOpt)
	}
	if err != nil {
		log.Print(err)
	}
	defer filterCursor.Close(ctx)

	var articles []Article
	if err = filterCursor.All(ctx, &articles); err != nil {
		log.Print(err)
	}

	json.NewEncoder(resp).Encode(articles)
}

func postArticles(resp http.ResponseWriter, req *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	resp.Header().Add("content-type", "application/json")

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
