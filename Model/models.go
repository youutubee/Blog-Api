package model

import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	Id           bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name         string        `json:"username" bson:"username"`
	RefreshToken string        `json:"-" bson:"refresh_token"`
	Password     string        `json:"password" bson:"password"`
}

type Blog struct {
	Id          bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title       string        `json:"title" bson:"title"`
	Description string        `json:"description" bson:"description"`
	BlogBody    string        `json:"blog_body" bson:"blog_body"`
	AuthorId    bson.ObjectID `json:"author_id" bson:"author_id"`
}
