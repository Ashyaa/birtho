package bot

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	U "github.com/ashyaa/birtho/util"
	"github.com/koffeinsource/go-imgur"
	"github.com/koffeinsource/go-klogger"
	LR "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Item struct {
	ID   int    `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
	Path string `json:"path" yaml:"path"`
	URL  string `json:"url" yaml:"url"`
}

type Config struct {
	AppID             int    `json:"app-id" yaml:"app-id"`
	ClientID          int    `json:"client-id" yaml:"client-id"`
	PublicKey         string `json:"public-key" yaml:"public-key"`
	ImgurClientID     string `json:"imgur-client-id" yaml:"imgur-client-id"`
	ImgurClientSecret string `json:"imgur-client-secret" yaml:"imgur-client-secret"`
	Token             string `json:"token" yaml:"token"`
	Items             []Item `json:"items" yaml:"items"`
	filepath          string
}

func ReadConfig(log LR.Logger) (Config, error) {
	filepath := os.Getenv("BIRTHO_CONFIG")
	if filepath == "" {
		filepath = "config/data.yml"
	}
	rawFile, err := os.Open(filepath)
	if err != nil {
		return Config{}, err
	}
	defer rawFile.Close()

	bytes, _ := ioutil.ReadAll(rawFile)
	var conf Config
	err = yaml.Unmarshal(bytes, &conf)
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

	for idx, item := range c.Items {
		if item.ID <= 0 {
			c.Items[idx].ID = max + 1
			max += 1
			log.Infof("wrote ID %d to item %s", c.Items[idx].ID, item.Name)
			modified = true
		}
		if item.URL == "" {
			filepath := path.Join(configDir, item.Path)
			img, st, err := client.UploadImageFromFile(filepath, "", "gothella", "")
			if st != 200 || err != nil {
				log.Errorf("failed to upload %s image to imgur: %v\n", item.Name, st)
				log.Errorf("failed to upload %s image to imgur: %v\n", item.Name, err)
			} else {
				log.Infof("succesfully uploaded %s image to imgur", item.Name)
				c.Items[idx].URL = img.Link
				modified = true
			}
		}
	}

	if modified {
		file, err := yaml.Marshal(c)
		if err != nil {
			return
		}
		ioutil.WriteFile(c.filepath, file, 0644)
	}
}
