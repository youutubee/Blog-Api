package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	model "goblogapi/Model"
)

const connectionstring = "mongodb+srv://justforyouutubee_db_user:bGRczHBexaycEKQe@cluster0.3jpfluu.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
const dbname = "BlogApp"
const coloumnname = "Blogs"

var jwtKey = []byte("secretkey")

var collection *mongo.Collection

func init() {
	clientOptions := options.Client().ApplyURI(connectionstring)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database(dbname).Collection(coloumnname)
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Function to check if the user is valid
func checkValidUser(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return false // add the code here to check and refresh the token
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return false
		}
	}

	tokenStr := cookie.Value

	claim := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claim, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if !tkn.Valid {
		// Check refresh token and give a new one
	}

	return true
}

// Function to generate a refresh token
func generateRefreshToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": username,
		"exp":      time.Now().Add(24 * time.Hour * 7).Unix(), // 7 days
	})
	return token.SignedString(jwtKey)
}

func hashUserPassword(string password) string{
	
}

// Controller function to create a new user if one dosent exists
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	err = collection.FindOne(context.TODO(), bson.M{"username": user.Name}).Err()
	if err == nil {
		log.Fatal("User already exists please login instead of signup")
	}
	if err != mongo.ErrNoDocuments {
		log.Fatal(err)
	}
	// if we reached till here that means that user dosent exists and there is no error also

	// use brycpt to hash password here afterwards checking


	_, err =collection.InsertOne(context.Background(), user)

	if err!=nil{
		log.Fatal("internal server error cant insert the user in db")
	}

	json.NewEncoder(w).Encode(user)
}

// Controllerfunctions to insert the blog
func InsertOneBlog(w http.ResponseWriter, r *http.Request) {
	// 1-> need to verify if that user can insert (writing a middleware for it)
	if !checkValidUser(w,r){
		log.Fatal("User not restricted to write a blog")
	}
	var blog model.Blog
	err := json.NewDecoder(r.Body).Decode(blog)
	if err!=nil{
		http.Error(w, "Invalid blog data", http.StatusBadRequest)
        return
	}
	_ , err = collection.InsertOne(context.Background(),blog)
	if err!=nil{
		http.Error(w,"Failed to insert the blog",http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(blog)
}

func LoginOneUser(w http.ResponseWriter , r *http.Request){

}
