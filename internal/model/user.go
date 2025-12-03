package model

import "time"

type JWTUserIDKeyType string
const JWTUserIDKey JWTUserIDKeyType = "user_id"

type User struct {
    ID        int64     `json:"id"`
	Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
}

type SignupPayload struct {
	Name     string `json:"name" required:"true" maxlen:"25"`
	Email    string `json:"email" required:"true" maxlen:"30"`
	Password string `json:"password" required:"true"`
}