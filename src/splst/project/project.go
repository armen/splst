package project

import (
	"github.com/armen/hdis"
	"github.com/garyburd/redigo/redis"
	"github.com/nfnt/resize"

	"bytes"
	"encoding/gob"
	"errors"
	"image"
	"image/draw"
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
	redisPool            *redis.Pool
)

func init() {
	redisPool = &redis.Pool{
		MaxIdle:     20,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
}

type Project struct {
	Id            string
	OwnerId       string
	URL           string
	Name          string
	Description   string
	RepositoryURL string
	Thumb         bool
}

func (p *Project) Save(appRoot string) error {

	_, err := url.ParseRequestURI(p.URL)
	if err != nil {
		return InvalidUrlError
	}

	c := redisPool.Get()
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

	err = p.generateThumbnail(appRoot)
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

func (p *Project) generateThumbnail(appRoot string) (err error) {

	projectsPath := path.Join(appRoot, "static", "projects")
	imgPath := path.Join(projectsPath, p.OwnerId, p.Id)
	err = os.MkdirAll(imgPath, 0777)
	if err != nil {
		return err
	}

	t := time.Now()
	output, err := exec.Command(path.Join(appRoot, "utils", "fetch-image.sh"), p.URL, path.Join(imgPath, "big.jpg")).CombinedOutput()
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

	// Final image should be 298px width
	resizedImg := resize.Resize(299, 0, img, resize.Lanczos3)
	out, _ := os.Create(path.Join(imgPath, "small.jpg"))
	defer out.Close()

	// remove left and top, 1px border
	rect := image.Rect(0, 0, 298, 174)
	thumb := image.NewRGBA(rect)
	draw.Draw(thumb, rect, resizedImg, image.Point{1, 1}, draw.Src)

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

func HasList(userid string) bool {

	c := redisPool.Get()
	defer c.Close()

	length, _ := redis.Int(c.Do("LLEN", "u:"+userid))
	return length > 0
}

func MyList(userid string) (*[]*Project, error) {
	return projectsList("u:"+userid, 0, 50)
}

func RecentList() (*[]*Project, error) {
	return projectsList("recent-projects", 0, 50)
}

func projectsList(key string, from, to int) (*[]*Project, error) {

	c := redisPool.Get()
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

func (p *Project) Delete(appRoot string) error {

	projectsPath := path.Join(appRoot, "static", "projects")
	imgPath := path.Join(projectsPath, p.OwnerId, p.Id)
	err := os.RemoveAll(imgPath)
	if err != nil {
		return err
	}

	// Delete the owner directory if it's empty, so any errors should be ignored
	ownerPath := path.Join(projectsPath, p.OwnerId)
	os.Remove(ownerPath)

	c := redisPool.Get()
	defer c.Close()

	hc := hdis.Conn{c}

	_, err = c.Do("LREM", "u:"+p.OwnerId, 1, p.Id)
	if err != nil {
		return err
	}

	_, err = c.Do("LREM", "recent-projects", 1, p.Id)
	if err != nil {
		return err
	}

	key := "p:" + p.Id
	if deleted, _ := redis.Bool(hc.Do("HDEL", key)); deleted {
		return nil
	}

	return ProjectDeletionError
}

func Fetch(pId string) (*Project, error) {

	c := redisPool.Get()
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
