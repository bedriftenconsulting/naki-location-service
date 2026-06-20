package main

import (
	"fmt"
	"log"

	"github.com/naki/location-service/conf"
	"github.com/naki/location-service/database"
	"github.com/naki/location-service/functions/api_functions"
	"github.com/naki/location-service/routers"
)

func main() {
	if err := conf.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	api_functions.InitRedis()

	api_functions.StartKafkaConsumers()

	r := routers.SetupRouter()

	addr := fmt.Sprintf(":%s", conf.AppConfig.Port)
	log.Printf("location service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
