package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var uri string
var mongoClient *mongo.Client

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri = os.Getenv("MONGODB_URI")
	fmt.Println("mongo URI is", uri)
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	if err := connect_to_mongodb(); err != nil {
		log.Fatal("Could not connect to MongoDB")
	}
	fmt.Println("Connected to MongoDB!", mongoClient)
}

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.POST("/albums", postAlbums)
	router.GET("/albums/:id", getAlbumByID)

	router.Run(":8000")
}

func connect_to_mongodb() error {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	mongoClient = client
	return err
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	// Find movies
	cursor, err := mongoClient.Database(os.Getenv("MONGO_DEFAULT_DATABASE")).Collection("movies").Find(context.TODO(), bson.D{{}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map results
	var movies []bson.M
	if err = cursor.All(context.TODO(), &movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return movies
	c.JSON(http.StatusOK, movies)
	// c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	fmt.Println("New album is", newAlbum)
	cursor, err := mongoClient.Database(os.Getenv("MONGO_DEFAULT_DATABASE")).Collection("movies").InsertMany(context.TODO(), []interface{}{newAlbum})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Inserted a new album with ID: ", cursor.InsertedIDs[0])
	c.IndentedJSON(http.StatusCreated, gin.H{"id": cursor.InsertedIDs[0]})
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")
	fmt.Println("id is", reflect.TypeOf(id))
	// Find album
	var album album
	err := mongoClient.Database(os.Getenv("MONGO_DEFAULT_DATABASE")).Collection("movies").FindOne(context.TODO(), bson.M{"id": id}).Decode(&album)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, album)
}
