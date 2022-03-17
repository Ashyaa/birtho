package bot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	U "github.com/ashyaa/birtho/util"
	"github.com/koffeinsource/go-imgur"
	"github.com/koffeinsource/go-klogger"
	LR "github.com/sirupsen/logrus"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
}

type Config struct {
	AppID             int    `json:"app-id"`
	ClientID          int    `json:"client-id"`
	PublicKey         string `json:"public-key"`
	ImgurClientID     string `json:"imgur-client-id"`
	ImgurClientSecret string `json:"imgur-client-secret"`
	Token             string `json:"token"`
	Items             []Item `json:"items"`
	filepath          string
}

func ReadConfig(log LR.Logger) (Config, error) {
	filepath := os.Getenv("BIRTHO_CONFIG")
	if filepath == "" {
		filepath = "config/data.json"
	}
	jsonFile, err := os.Open(filepath)
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
	conf.filepath = filepath
	conf.initImages(log)
	return conf, nil
}

func (c *Config) initImages(log LR.Logger) {
	modified := false
	IDs := []int{}
	for _, img := range c.Items {
		IDs = append(IDs, img.ID)
	}
	max := U.Max(IDs)

	client := new(imgur.Client)
	client.HTTPClient = new(http.Client)
	client.Log = new(klogger.CLILogger)
	client.ImgurClientID = *&c.ImgurClientID
	configDir := path.Dir(c.filepath)

	for idx, img := range c.Items {
		if img.ID <= 0 {
			c.Items[idx].ID = max + 1
			max += 1
			log.Info("wrote ID %d to item %s", c.Items[idx].ID, img.Name)
			modified = true
		}
		if img.URL == "" {
			filepath := path.Join(configDir, img.Path)
			img, st, err := client.UploadImageFromFile(filepath, "", "gothella", "")
			if st != 200 || err != nil {
				log.Errorf("failed to upload %s image to imgur: %v\n", img.Name, st)
				log.Errorf("failed to upload %s image to imgur: %v\n", img.Name, err)
			} else {
				log.Infof("succesfully uploaded %s image to imgur", img.Name)
				c.Items[idx].URL = img.Link
				modified = true
			}
		}
	}

	if modified {
		file, err := json.MarshalIndent(c, "", "    ")
		if err != nil {
			return
		}
		ioutil.WriteFile(c.filepath, file, 0644)
	}
}
