package project

import (
	"github.com/garyburd/redigo/redis"

	"bytes"
	"encoding/gob"
	"errors"
	"net/url"
)

type Project struct {
	Name    string
	URL     string
	OwnerId string
}

func (p *Project) Save() error {

	url, err := url.ParseRequestURI(p.URL)
	if err != nil {
		return err
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

	key := "p-" + url.Host
	exists, err := redis.Bool(c.Do("EXISTS", key))
	if err != nil {
		return err
	}

	if exists {
		return errors.New("The project already exists")
	}

	_, err = c.Do("SET", key, buffer.Bytes())

	return err
}
