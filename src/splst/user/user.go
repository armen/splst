package user

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	mc = memcache.New("127.0.0.1:11211")
)

type User struct {
	Userid       uint64
	RegisteredAt time.Time
	LastAccess   time.Time
	Name         string
	Email        string
}

func New(userid uint64) *User {
	return &User{Userid: userid, RegisteredAt: time.Now()}
}

func (user *User) Save() error {

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(user)
	if err != nil {
		return err
	}

	if user.Userid != 0 {
		err := mc.Set(&memcache.Item{Key: strconv.FormatUint(user.Userid, 10), Value: buffer.Bytes()})
		return err
	}

	return errors.New("Ignored saving user because userid is 0")
}

func (user *User) Fetch() error {

	u, err := mc.Get(strconv.FormatUint(user.Userid, 10))
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(u.Value))
	err = dec.Decode(&user)

	return err
}

func (user *User) Update() error {

	u, err := mc.Get(strconv.FormatUint(user.Userid, 10))
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(user)
	if err != nil {
		return err
	}

	u.Value = buffer.Bytes()
	err = mc.CompareAndSwap(u)

	return err
}
