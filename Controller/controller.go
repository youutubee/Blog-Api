package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"github.com/joho/godotenv"

	model "goblogapi/Model"
)

var jwtKey = []byte("secretkey")

var collection *mongo.Collection

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connectionstring := os.Getenv("MONGO_URI")
	dbname := os.Getenv("DBNAME")
	coloumnname := os.Getenv("COLNAME")
	jwtKey = []byte(os.Getenv("JWT_SECRET"))

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
			return false
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
		refreshcookie, err := r.Cookie("refresh_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized) // no refresh token, user must log in again
			return false
		}

		refreshTokenStr := refreshcookie.Value

		claims := &Claims{}

		rtkn, err := jwt.ParseWithClaims(refreshTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !rtkn.Valid {
			w.WriteHeader(http.StatusUnauthorized) // refresh token expired/invalid
			return false
		}

		err = collection.FindOne(
			context.Background(),
			bson.M{"username": claims.Username, "refresh_token": refreshTokenStr},
		).Err()

		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusUnauthorized) // refresh token not found in DB
			return false
		}

		// if we reach here that means the refresh token was found and its not expiry
		newAccessToken, _ := generateAccessToken(claims.Username)
		newRefreshToken, _ := generateRefreshToken(claims.Username)

		// setting the active token as a cookie for the user
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    newAccessToken,
			Expires:  time.Now().Add(15 * time.Minute),
			HttpOnly: true,
		})

		// setting the refresh token also in the cookie for the user
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken, // only if you’re rotating refresh tokens
			Expires:  time.Now().Add(24 * 7 * time.Hour),
			HttpOnly: true,
		})

		_, err = collection.UpdateOne(context.Background(), bson.M{"username": claims.Username}, bson.M{"$set": bson.M{"refresh_token": newRefreshToken}})
		if err != nil {
			http.Error(w, "Cannot update the refresh token when it was valid", http.StatusInternalServerError)
			return false
		}
	}
	return true
}

// Function to generate an access (active) token
func generateAccessToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(), // expires in 15 minutes
	})
	return token.SignedString(jwtKey)
}

// Function to generate a refresh token
func generateRefreshToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": username,
		"exp":      time.Now().Add(24 * time.Hour * 7).Unix(), // 7 days
	})
	return token.SignedString(jwtKey)
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

	_, err = collection.InsertOne(context.Background(), user)

	if err != nil {
		log.Fatal("internal server error cant insert the user in db")
	}

	json.NewEncoder(w).Encode(user)
}

// Controllerfunctions to insert the blog
func InsertOneBlog(w http.ResponseWriter, r *http.Request) {
	// 1-> need to verify if that user can insert (writing a middleware for it)
	if !checkValidUser(w, r) {
		log.Fatal("User not restricted to write a blog")
	}
	var blog model.Blog
	err := json.NewDecoder(r.Body).Decode(&blog)
	if err != nil {
		http.Error(w, "Invalid blog data", http.StatusBadRequest)
		return
	}
	_, err = collection.InsertOne(context.Background(), blog)
	if err != nil {
		http.Error(w, "Failed to insert the blog", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(blog)
}

func LoginOneUser(w http.ResponseWriter, r *http.Request) {
	var user model.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Cannot decdoe the json error", http.StatusBadRequest)
	}

	var foundUser model.User

	err = collection.FindOne(context.Background(), bson.M{"username": user.Name}).Decode(&foundUser)

	if err == mongo.ErrNoDocuments {
		http.Error(w, "User not found or password is not correct", http.StatusUnauthorized)
	}

	if err != nil {
		http.Error(w, "Some other error", http.StatusInternalServerError)
	}

	// if we reached till here that means the user exists and we need to check the password now
	if foundUser.Password != user.Password {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// if reached here the username and password both are correct and now we need to create the refersh token and active token
	accessToken, _ := generateAccessToken(user.Name)
	refreshToken, _ := generateRefreshToken(user.Name)

	// Store the refresh token in the db
	_, err = collection.UpdateOne(context.Background(), bson.M{"username": user.Name}, bson.M{"$set": bson.M{"refresh_token": refreshToken}}) // we use set to set the value of the refresh token cos at the first time it could be empty
	if err != nil {
		http.Error(w, "Cannot update or insert the refresh token in the db", http.StatusInternalServerError)
		return
	}

	// Send tokens back to client
	http.SetCookie(w, &http.Cookie{
		Name:     "token", // access token
		Value:    accessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(24 * 7 * time.Hour),
		HttpOnly: true,
	})

	json.NewEncoder(w).Encode(user)
}
