package main

type User struct {
    id int64 `json:"user_id"`
    name string `json:"user_name"`
    roles []string `json:"roles"`
}