package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type IrtData struct {
	Subjects map[string]SubjectData `json:"subjects"`
}

type SubjectData struct {
	Task1 int `json:"task1"`
	Task2 int `json:"task2"`
	Task3 int `json:"task3"`
}

type IrtResponse struct {
	Abilities        map[string]float32 `json:"abilities"`
	Difficult        map[string]float32 `json:"difficult"`
	Err              float32            `json:"err"`
	RejectedTasks    []string           `json:"rejected_tasks"`
	RejectedSubjects []string           `json:"rejected_subjects"`
}

type Config struct {
	urlAPI    string
	urlTest   string
	redisHost string
	redisPort string
}

func redisClient(redisHost string, redisPort string, redisPassword string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0,
	})
	return client
}

func readConf() (Config, error) {
	err := godotenv.Load()

	var config Config

	if err != nil {
		return config, err
	}

	config.urlTest = os.Getenv("urlTest")
	config.urlAPI = os.Getenv("urlAPI")

	config.redisHost = os.Getenv("REDIS_HOST")
	config.redisPort = os.Getenv("REDIS_PORT")

	return config, nil
}

func irtData() (IrtData, error) {
	dataMap := make(map[string]SubjectData)

	subj1 := SubjectData{Task1: 1, Task2: 0, Task3: 0}
	subj2 := SubjectData{Task1: 0, Task2: 1, Task3: 1}
	subj3 := SubjectData{Task1: 1, Task2: 1, Task3: 0}

	dataMap["subj1"] = subj1
	dataMap["subj2"] = subj2
	dataMap["subj3"] = subj3

	irtData := IrtData{Subjects: dataMap}

	return irtData, nil
}

func requestIrt(url string, irtData []byte) IrtResponse {

	req, err := http.NewRequest("POST", url, bytes.NewReader(irtData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var irtResp IrtResponse

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &irtResp)

	fmt.Printf("response Body:", irtResp)

	return irtResp
}

func main() {
	log.Printf("Client started!")

	config, err := readConf()
	if err != nil {
		log.Fatalf("Failed to read config: \n%s", err)
	}
	rds := redisClient(
		config.redisHost,
		config.redisPort,
		os.Getenv("REDIS_PASSWORD"))

	ping, err := rds.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to ping redis service: \n%s", err)
	}
	log.Printf("Redis ping:%s", ping)

	url := config.urlAPI
	log.Printf("target url: %s", url)

	irt, err := irtData()
	if err != nil {
		log.Fatalf("Failed to read irt data: \n%s", err)
	}

	jsonData, err := json.Marshal(irt)

	irtResult := requestIrt(url, jsonData)
	log.Printf("result: %v", irtResult)
}
