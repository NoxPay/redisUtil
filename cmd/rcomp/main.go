package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"crg.eti.br/go/config"
	_ "crg.eti.br/go/config/ini"
	"github.com/go-redis/redis/v8"
)

type Config struct {
	RedisAddress string `json:"redis_address" ini:"redis_address" cfg:"redis_address" cfgDefault:"localhost:6379"`
	RedisPass    string `json:"redis_password" ini:"redis_password" cfg:"redis_password" cfgDefault:""`
	RedisUser    string `json:"redis_user" ini:"redis_user" cfg:"redis_user" cfgDefault:""`
	RedisFilter  string `json:"redis_filter" ini:"redis_filter" cfg:"redis_filter" cfgDefault:"*"`
	RedisDB      int    `json:"redis_db" ini:"redis_db" cfg:"redis_db" cfgDefault:"0"`
	OutputFile   string `json:"output_file" ini:"output_file" cfg:"output_file" cfgDefault:"data.json"`
}

func main() {
	cfg := Config{}
	config.File = "config.ini"
	err := config.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPass,
		Username: cfg.RedisUser,
		DB:       cfg.RedisDB,
	})

	jsonData, err := os.ReadFile(cfg.OutputFile)
	if err != nil {
		log.Fatal(err)
	}

	var jsonDataMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonDataMap)
	if err != nil {
		log.Fatal(err)
	}

	keys, err := client.Keys(context.Background(), cfg.RedisFilter).Result()
	if err != nil {
		log.Fatal(err)
	}

	for key, val := range jsonDataMap {
		redisVal, _ := client.Get(context.Background(), key).Result()
		if redisVal != val {
			fmt.Printf("Different value: %q\n\tRedis = %q\n\tJSON  = %q\n", key, redisVal, val)
		}
	}

	for _, key := range keys {
		if _, ok := jsonDataMap[key]; !ok {
			fmt.Printf("Missing key in JSON: %q\n", key)
		}
	}

	for key := range jsonDataMap {
		if client.Exists(context.Background(), key).Val() == 0 {
			fmt.Printf("Missing key in Redis: %q\n", key)
		}
	}
}
