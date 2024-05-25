package main

import (
	// "github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	hostAPI string
}

func readConf() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file: \n%s", err)
	}
	hostAPI := os.Getenv("hostAPI")
	config := Config{hostAPI}
	return config
}

func main() {
	log.Printf("Client started!")
	config := readConf()
	log.Printf("target host: %s", config.hostAPI)
	file, err := os.Open("data/test.xlsx")
	if err != nil {
		log.Fatal("Can't open file!")
	}

	log.Printf("%v", file)
}
