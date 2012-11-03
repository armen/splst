package project

import (
	"github.com/garyburd/redigo/redis"

	"bytes"
	"encoding/gob"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path"
	"splst/utils"
)

var (
	InvalidUrlError    = errors.New("invalid URL")
	GenerateThumbError = errors.New("couldn't generate thumbnail image")
)

type Project struct {
	Id      string
	Name    string
	URL     string
	OwnerId string
}

func (p *Project) Save() error {

	_, err := url.ParseRequestURI(p.URL)
	if err != nil {
		return InvalidUrlError
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(p)
	if err != nil {
		return err
	}

	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return err
	}
	defer c.Close()

	// want to use a short project id and should be double checked for existance
	var key string

	for {
		p.Id = utils.GenId(3)
		key = "p-" + p.Id
		exists, _ := redis.Bool(c.Do("EXISTS", key))

		if !exists {
			break
		}
	}

	_, err = c.Do("SET", key, buffer.Bytes())
	if err != nil {
		return err
	}

	_, err = c.Do("RPUSH", "u-"+p.OwnerId, p.Id)
	if err != nil {
		return err
	}

	_, err = c.Do("LPUSH", "recent-projects", p.Id)
	if err != nil {
		return err
	}

	return err
}

func (p *Project) GenerateThumbnail(rootPath string) error {

	imgPath := path.Join(rootPath, p.OwnerId, p.Id)
	err := os.MkdirAll(imgPath, 0777)
	if err != nil {
		return err
	}

	err = exec.Command("wkhtmltoimage-amd64", p.URL, path.Join(imgPath, "big.png")).Run()
	if err != nil {
		return GenerateThumbError
	}

	return nil
}
