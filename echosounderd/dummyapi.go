package echosounderd

import (
	"encoding/json"
	"net/http"
	"time"
)

type post struct {
	ID     int    `json:"ID"`
	UserID int    `json:"userId"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func requestDummyAPI() (post, error) {
	var p post

	// TODO: configurable
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		return p, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&p); err != nil {
		return p, err
	}

	return p, nil
}
