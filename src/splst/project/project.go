package project

import (
	"github.com/armen/hdis"
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
	"time"
)

var (
	InvalidUrlError      = errors.New("invalid URL")
	GenerateThumbError   = errors.New("couldn't generate thumbnail image")
	ProjectDeletionError = errors.New("couldn't delete the project")
)

type Project struct {
	Id            string
	OwnerId       string
	URL           string
	Name          string
	Description   string
	RepositoryURL string
	Thumb         bool
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

	hc := hdis.Conn{c}

	for {
		p.Id = utils.GenId(3)
		key = "p:" + p.Id

		if exists, _ := redis.Bool(hc.Do("HEXISTS", key)); !exists {
			break
		}
	}

	err = p.generateThumbnail(rootPath)
	if err != nil {
		log.Println(err)
	}
	p.Thumb = err == nil

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(p)
	if err != nil {
		return err
	}

	_, err = hc.Set(key, buffer.Bytes())
	if err != nil {
		return err
	}

	_, err = c.Do("RPUSH", "u:"+p.OwnerId, p.Id)
	if err != nil {
		return err
	}

	_, err = c.Do("LPUSH", "recent-projects", p.Id)
	if err != nil {
		return err
	}

	return err
}

func (p *Project) generateThumbnail(rootPath string) (err error) {

	imgPath := path.Join(rootPath, p.OwnerId, p.Id)
	err = os.MkdirAll(imgPath, 0777)
	if err != nil {
		return err
	}

	t := time.Now()
	output, err := exec.Command(path.Join(os.Getenv("PWD"), "utils", "fetch-image.sh"), p.URL, path.Join(imgPath, "big.jpg")).CombinedOutput()
	if err != nil {
		log.Println(err, string(output))
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}

	// Check the timeout
	if time.Since(t).Seconds() > 90 {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}

	f, err := os.Open(path.Join(imgPath, "big.jpg"))
	if err != nil {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}

	// Decode the whole file
	img, _, err := image.Decode(f)
	if err != nil {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}

	thumb := resize.Resize(318, 0, img, resize.Lanczos3)
	out, _ := os.Create(path.Join(imgPath, "small.jpg"))
	defer out.Close()

	err = jpeg.Encode(out, thumb, &jpeg.Options{Quality: 90})
	if err != nil {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}

	return nil
}

func (project Project) Mine(userid interface{}) bool {
	return project.OwnerId == userid.(string)
}

func MyList(userid string) (*[]*Project, error) {
	return projectsList("u:"+userid, 0, 50)
}

func RecentList() (*[]*Project, error) {
	return projectsList("recent-projects", 0, 50)
}

func projectsList(key string, from, to int) (*[]*Project, error) {

	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	recentList, err := redis.Values(c.Do("LRANGE", key, from, to))
	if err != nil {
		return nil, err
	}

	var projects []*Project

	for len(recentList) > 0 {
		var pId string

		recentList, err = redis.Scan(recentList, &pId)
		if err != nil {
			return nil, err
		}

		project, err := Fetch(pId)
		if err != nil {
			log.Println(err)
		}
		projects = append(projects, project)
	}

	return &projects, nil
}

func (p *Project) Delete(rootPath string) error {

	imgPath := path.Join(rootPath, p.OwnerId, p.Id)
	err := os.RemoveAll(imgPath)
	if err != nil {
		return err
	}

	// Delete the owner directory if it's empty, so any errors should be ignored
	ownerPath := path.Join(rootPath, p.OwnerId)
	os.Remove(ownerPath)

	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return err
	}
	defer c.Close()

	hc := hdis.Conn{c}

	key := "p:" + p.Id

	if deleted, _ := redis.Bool(hc.Do("HDEL", key)); deleted {
		return nil
	}

	return ProjectDeletionError
}

func Fetch(pId string) (*Project, error) {

	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var project Project

	hc := hdis.Conn{c}

	p, err := redis.Bytes(hc.Get("p:" + pId))
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buffer)
	dec.Decode(&project)

	return &project, nil
}
