package main

import (
	"io"
	"log"

	"github.com/go-redis/redis"
)

func logClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Println(err)
	}
}

func main() {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer logClose(client)

	if result, err := client.Ping().Result(); err != nil {
		log.Fatal(result, err)
	}

	result, err := client.Get("key").Result()
	log.Println(result, err)
}
