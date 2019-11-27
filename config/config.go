package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Account struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	Reciver Account
	Sender  Account
	User    Account
)

func init() {
	a := make(map[string]Account)

	data, err := ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &a)
	if err != nil {
		log.Fatal(err)
	}

	Reciver = a["reciver"]
	Sender = a["sender"]
	User = a["user"]
}

func ReadFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return data, nil
}
