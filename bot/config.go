package bot

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	AppID     int    `json:"app-id"`
	ClientID  int    `json:"client-id"`
	PublicKey string `json:"public-key"`
	Token     string `json:"token"`
}

func ReadConfig() (Config, error) {
	jsonFile, err := os.Open("config/data.json")
	if err != nil {
		return Config{}, err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)
	var conf Config
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}
