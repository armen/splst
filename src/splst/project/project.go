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
	InvalidUrlError    = errors.New("invalid URL")
	GenerateThumbError = errors.New("couldn't generate thumbnail image")
	ProjectDeleteError = errors.New("couldn't delete the project")
	ProjectSaveError   = errors.New("couldn't save the project")
	redisPool          *redis.Pool
	appRoot            string
	saveQueue          chan job
)

type job struct {
	project *Project
	err     chan error
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

func (p *Project) save() error {

	c := redisPool.Get()
	defer c.Close()

	hc := hdis.Conn{c}

	// Decrease number of jobs and remove it from the queue
	defer hc.Do("HINCRBY", "u:"+p.OwnerId+":jobs", -1)
	defer c.Do("SREM", "jobs-in-queue", p.Id)

	log.Printf("Saving project %q, %q", p.Id, p.URL)

	err := p.generateThumbnail()
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

	key := "p:" + p.Id

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

	log.Printf("Saved project %q, %q", p.Id, p.URL)

	return nil
}

func (p *Project) Save() error {

	_, err := url.ParseRequestURI(p.URL)
	if err != nil {
		return InvalidUrlError
	}

	c := redisPool.Get()
	defer c.Close()

	hc := hdis.Conn{c}

	// Use a short project id and double checked for it's existance
	for {
		p.Id = utils.GenId(3)
		key := "p:" + p.Id

		if exists, _ := redis.Bool(hc.Do("HEXISTS", key)); !exists {
			break
		}
	}

	if ok, _ := redis.Bool(c.Do("SADD", "jobs-in-queue", p.Id)); ok {

		hc.Do("HINCRBY", "u:"+p.OwnerId+":jobs", 1)

		j := job{project: p, err: make(chan error)}

		// Put the job in the queue, this will block if it's full
		saveQueue <- j

		// Read the result from err channel
		return <-j.err
	}

	return ProjectSaveError
}

func (p *Project) generateThumbnail() (err error) {

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

	// final image should be 318x185
	resizedImg := resize.Resize(319, 0, img, resize.Lanczos3)
	smallFile, err := os.Create(path.Join(imgPath, "small.jpg"))
	if err != nil {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}
	defer smallFile.Close()

	// remove left and top, 1px border
	rect := image.Rect(0, 0, 318, 185)
	small := image.NewRGBA(rect)
	draw.Draw(small, rect, resizedImg, image.Point{1, 1}, draw.Src)

	err = jpeg.Encode(smallFile, small, &jpeg.Options{Quality: 90})
	if err != nil {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}

	// big sized image
	resizedImg = resize.Resize(429, 0, img, resize.Lanczos3)
	bigFile, err := os.Create(path.Join(imgPath, "big.jpg"))
	if err != nil {
		os.RemoveAll(imgPath)
		return GenerateThumbError
	}
	defer bigFile.Close()

	y := resizedImg.Bounds().Dy()
	if y > 415 {
		y = 415
	}

	// remove left and top, 1px border
	rect = image.Rect(0, 0, 428, y-1)
	big := image.NewRGBA(rect)
	draw.Draw(big, rect, resizedImg, image.Point{1, 1}, draw.Src)

	err = jpeg.Encode(bigFile, big, &jpeg.Options{Quality: 90})
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
	return ListCount(userid) > 0
}

func ListCount(userid string) int {
	c := redisPool.Get()
	defer c.Close()

	length, _ := redis.Int(c.Do("LLEN", "u:"+userid))
	return length
}

func JobsCount(userid string) int {
	c := redisPool.Get()
	defer c.Close()

	hc := hdis.Conn{c}
	count, _ := redis.Int(hc.Do("HGET", "u:"+userid+":jobs"))

	return count
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

func (p *Project) Delete() error {

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

	return ProjectDeleteError
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

func Init(pool *redis.Pool, appRoot string, saveConcurrencySize int) {
	redisPool = pool
	appRoot = appRoot
	saveQueue = make(chan job, saveConcurrencySize)

	for i := 0; i < saveConcurrencySize; i++ {
		// Create workers
		go func() {
			for {
				select {
				// Read a job from the queue, save the project and put the result in err channel
				case j := <-saveQueue:
					j.err <- j.project.save()
				}
			}
		}()
	}
}
