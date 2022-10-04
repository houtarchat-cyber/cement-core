package cement

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

func UserLogin(username string, password string) error {
	bucket := GetBucket()
	username = base64.StdEncoding.EncodeToString([]byte(username))
	password = fmt.Sprintf("%x", md5.Sum([]byte(password)))

	isExist, err := bucket.IsObjectExist("users/" + username)
	if err != nil {
		return err
	}
	if !isExist {
		return errors.New("user not exist")
	}
	pwdfile, err := bucket.GetObject("users/" + username)
	if err != nil {
		return err
	}

	defer pwdfile.Close()

	pwd, err := ioutil.ReadAll(pwdfile)
	if err != nil {
		return err
	}

	if string(pwd) != password {
		return errors.New("password error")
	}
	return nil
}

func UserCreate(username string, password string) error {
	bucket := GetBucket()
	username = base64.StdEncoding.EncodeToString([]byte(username))
	password = fmt.Sprintf("%x", md5.Sum([]byte(password)))

	isExist, err := bucket.IsObjectExist("users/" + username)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("user already exist")
	}

	err = bucket.PutObject("users/"+username, strings.NewReader(password))
	if err != nil {
		return err
	}

	err = bucket.PutObject("files/"+username+"/", strings.NewReader(""))
	if err != nil {
		return err
	}
	return nil
}
