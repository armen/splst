package project

import (
	"github.com/garyburd/redigo/redis"
	"github.com/nfnt/resize"

	"bytes"
	"encoding/gob"
	"errors"
	"image"
	"image/jpeg"
	"log"
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
	Id            string
	OwnerId       string
	URL           string
	Name          string
	Description   string
	RepositoryURL string
}

func (p *Project) Save(rootPath string) error {

	_, err := url.ParseRequestURI(p.URL)
	if err != nil {
		return InvalidUrlError
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

	err = p.generateThumbnail(rootPath)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(p)
	if err != nil {
		return err
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

func (p *Project) generateThumbnail(rootPath string) error {

	imgPath := path.Join(rootPath, p.OwnerId, p.Id)
	err := os.MkdirAll(imgPath, 0777)
	if err != nil {
		return err
	}

	output, err := exec.Command(path.Join(os.Getenv("PWD"), "utils", "fetch-image.sh"), p.URL, path.Join(imgPath, "big.jpg")).CombinedOutput()
	if err != nil {
		log.Println(err, string(output))
		return GenerateThumbError
	}

	f, err := os.Open(path.Join(imgPath, "big.jpg"))
	if err != nil {
		return GenerateThumbError
	}

	// Decode the whole file
	img, _, err := image.Decode(f)
	if err != nil {
		return GenerateThumbError
	}

	thumb := resize.Resize(228, 0, img, resize.Lanczos3)
	out, _ := os.Create(path.Join(imgPath, "small.jpg"))
	defer out.Close()

	err = jpeg.Encode(out, thumb, &jpeg.Options{Quality: 90})
	if err != nil {
		return GenerateThumbError
	}

	return nil
}

func RecentList() (*[]Project, error) {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	recentList, err := redis.Values(c.Do("LRANGE", "recent-projects", 0, 50))
	if err != nil {
		return nil, err
	}

	var projects []Project
	var project Project

	for len(recentList) > 0 {
		var pid string
		recentList, err = redis.Scan(recentList, &pid)
		if err != nil {
			return nil, err
		}

		p, err := redis.Bytes(c.Do("GET", "p-"+pid))
		if err != nil {
			return nil, err
		}

		buffer := bytes.NewBuffer(p)
		dec := gob.NewDecoder(buffer)
		dec.Decode(&project)

		projects = append(projects, project)
	}

	return &projects, nil
}
