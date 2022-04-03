package bot

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"

	U "github.com/ashyaa/birtho/util"
	"github.com/koffeinsource/go-imgur"
	"github.com/koffeinsource/go-klogger"
	LR "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Range struct {
	min int
	max int
}

func (r Range) Belongs(n int) bool {
	return r.min <= n && n <= r.max
}

type Item struct {
	Name   string  `json:"name" yaml:"name"`
	Chance float64 `json:"chance" yaml:"chance"`
	Range  Range
}

type Monster struct {
	ID               int     `json:"id" yaml:"id"`
	Name             string  `json:"name" yaml:"name"`
	Path             string  `json:"path" yaml:"path"`
	URL              string  `json:"url" yaml:"url"`
	Chance           float64 `json:"chance" yaml:"chance"`
	Items            []Item  `json:"items" yaml:"items"`
	Range            Range
	EqualItemChances bool
}

// Build the monster data for the game. Returns false if a major error was encountered, else true.
func (m *Monster) buildItemRanges(log LR.Logger) bool {
	sum := 1
	if len(m.Items) == 0 {
		return false
	}
	for i, item := range m.Items {
		chance := int(item.Chance * 100)
		m.Items[i].Range.min = sum
		m.Items[i].Range.max = sum + chance - 1
		sum += chance
	}
	if sum-1 != 10000 {
		if sum > 1 {
			log.Warnf("the sum of item spawn chances for monster '%s' is not equal to 100", m.Name)
		}
		m.EqualItemChances = true
		log.Infof("all items for monster '%s' set to have equal chances to spawn", m.Name)
	}
	return true
}

func (m Monster) RandomItem(log LR.Logger) Item {
	if m.EqualItemChances {
		index := rand.Intn(len(m.Items))
		return m.Items[index]
	}
	number := rand.Intn(10000) + 1
	for _, item := range m.Items {
		if item.Range.Belongs(number) {
			return item
		}
	}
	log.Fatal("invalid item roll")
	return Item{}
}

type Config struct {
	AppID             int       `json:"app-id" yaml:"app-id"`
	ClientID          int       `json:"client-id" yaml:"client-id"`
	PublicKey         string    `json:"public-key" yaml:"public-key"`
	ImgurClientID     string    `json:"imgur-client-id" yaml:"imgur-client-id"`
	ImgurClientSecret string    `json:"imgur-client-secret" yaml:"imgur-client-secret"`
	Token             string    `json:"token" yaml:"token"`
	Monsters          []Monster `json:"monsters" yaml:"monsters"`
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
	conf.init(log)
	return conf, nil
}

func (c *Config) init(log LR.Logger) {
	modified := false
	IDs := []int{}
	for _, m := range c.Monsters {
		IDs = append(IDs, m.ID)
	}
	max := U.Max(IDs)

	client := new(imgur.Client)
	client.HTTPClient = new(http.Client)
	client.Log = new(klogger.CLILogger)
	client.ImgurClientID = *&c.ImgurClientID
	configDir := path.Dir(c.filepath)

	for idx, monster := range c.Monsters {
		if monster.ID <= 0 {
			c.Monsters[idx].ID = max + 1
			max += 1
			log.Infof("wrote ID %d to item %s", c.Monsters[idx].ID, monster.Name)
			modified = true
		}
		if monster.URL == "" {
			filepath := path.Join(configDir, monster.Path)
			img, st, err := client.UploadImageFromFile(filepath, "", "gothella", "")
			if st != 200 || err != nil {
				log.Errorf("failed to upload %s image to imgur: %v\n", monster.Name, st)
				log.Errorf("failed to upload %s image to imgur: %v\n", monster.Name, err)
			} else {
				log.Infof("succesfully uploaded %s image to imgur", monster.Name)
				c.Monsters[idx].URL = img.Link
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
