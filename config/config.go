package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type User struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	Reciver User
	Sender  User
)

func init() {
	a := make(map[string]User)

	data, err := readFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &a)
	if err != nil {
		log.Fatal(err)
	}

	Reciver = a["reciver"]
	Sender = a["sender"]
}

func readFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return data, nil
}
