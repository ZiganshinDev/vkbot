package app

import (
	"log"

	"github.com/ZiganshinDev/scheduleVKBot/internal/service"
	"github.com/joho/godotenv"
)

func Run() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file. %v", err)
	}

	service.CreateBot()
}
