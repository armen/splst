package project

import (
	"github.com/garyburd/redigo/redis"

	"bytes"
	"encoding/gob"
	"errors"
	"net/url"
	"splst/utils"
)

var InvalidUrlError = errors.New("invalid URL")

type Project struct {
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
	var pid string

	for {
		pid = utils.GenId(3)
		key = "p-" + pid
		exists, _ := redis.Bool(c.Do("EXISTS", key))

		if !exists {
			break
		}
	}

	_, err = c.Do("SET", key, buffer.Bytes())
	if err != nil {
		return err
	}

	_, err = c.Do("RPUSH", "u-"+p.OwnerId, pid)
	if err != nil {
		return err
	}

	_, err = c.Do("LPUSH", "recent-projects", pid)
	if err != nil {
		return err
	}

	return err
}
