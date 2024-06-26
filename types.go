package main

import (
	"encoding/json"
	"time"
)

type Basket struct {
	ID        int             `json:"id"`
	Username  string          `json:"username"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Data      json.RawMessage `json:"data"`
	State     int             `json:"state"`
}
type CreateBasket struct {
	Username string          `json:"username"`
	State    int             `json:"state"`
	Data     json.RawMessage `json:"data"`
}
type GetBasket struct {
	ID        int             `json:"id"`
	Username  string          `json:"username"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Data      json.RawMessage `json:"data"`
	State     int             `json:"state"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
