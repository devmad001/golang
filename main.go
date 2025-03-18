package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
    ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name  string            `json:"name" bson:"name"`
    Email string            `json:"email" bson:"email"`
}

var client *mongo.Client
var collection *mongo.Collection

func init() {
    // Set up MongoDB connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Update the connection string according to your MongoDB setup
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    var err error
    
    client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Ping the database to verify connection
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")

    // Get collection handle
    collection = client.Database("userdb").Collection("users")

    // Create unique index for email
    indexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    _, err = collection.Indexes().CreateOne(ctx, indexModel)
    if err != nil {
        log.Fatal(err)
    }
}

func homePage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to Go Web Server with MongoDB!")
}

func createUser(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := collection.InsertOne(ctx, user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    user.ID = result.InsertedID.(primitive.ObjectID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var users []User
    if err = cursor.All(ctx, &users); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func main() {
    defer func() {
        if client != nil {
            if err := client.Disconnect(context.Background()); err != nil {
                log.Printf("Error disconnecting from MongoDB: %v\n", err)
            }
        }
    }()

    http.HandleFunc("/", homePage)
    http.HandleFunc("/users", createUser)
    http.HandleFunc("/users/list", getUsers)

    fmt.Println("Starting web server on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
} 