package main

import (
	"bytes"
	"context"
	// "database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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
	psqlHost  string
	psqlPort  string
	psqlDB    string
}

func redisClient(redisHost string, redisPort string, redisPassword string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0,
	})
	return client
}

func postgresDSN(psqlUser string, psqlPassword string, psqlHost string, psqlPort string, database string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", psqlUser, psqlPassword, psqlHost, psqlPort, database)
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

	config.psqlHost = os.Getenv("POSTGRES_HOST")
	config.psqlPort = os.Getenv("POSTGRES_PORT")
	config.psqlDB = os.Getenv("POSTGRES_DATABASE_NAME")

	return config, nil
}

func irtData(testId int) (IrtData, error) {
	log.Printf("get irt-data for %d", testId)
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

func ProcessIRT(testId int, url string, rds *redis.Client) {
	// get IRT data from some external source
	irt, err := irtData(testId)
	if err != nil {
		log.Fatalf("Failed to read irt data: \n%s", err)
	}

	jsonData, err := json.Marshal(irt)
	irtResult := requestIrt(url, jsonData)
	jsonIrt, err := json.Marshal(irtResult)
	if err != nil {
		log.Fatalf("can't serialize irt request: %s", err)
	}

	// caching
	err = rds.Set(context.Background(), strconv.Itoa(testId), jsonIrt, 0).Err()
	if err != nil {
		log.Fatalf("can't save data to cache: %s", err)
	}

	log.Printf("result: %v", irtResult)
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

	sqlDSN := postgresDSN(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), config.psqlHost, config.psqlPort, config.psqlDB)
	log.Println(sqlDSN)

	url := config.urlAPI
	log.Printf("target url: %s", url)

	testId := 0

	val, err := rds.Get(context.Background(), strconv.Itoa(testId)).Result()
	if val == "" {
		ProcessIRT(testId, config.urlAPI, rds)
	} else {
		log.Printf("IRT result is %v", val)
	}

}
