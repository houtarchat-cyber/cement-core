package cement

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

func UserLogin(username string, password string) error {
	bucket := getBucket()
	username = base64.StdEncoding.EncodeToString([]byte(username))
	password = fmt.Sprintf("%x", md5.Sum([]byte(password)))

	isExist, err := bucket.IsObjectExist("users/" + username)
	if err != nil {
		return err
	}
	if !isExist {
		return errors.New("user not exist")
	}
	passwordFile, err := bucket.GetObject("users/" + username)
	if err != nil {
		return err
	}

	defer func(passwordFile io.ReadCloser, err *error) {
		*err = passwordFile.Close()
	}(passwordFile, &err)
	if err != nil {
		return err
	}

	pwd, err := io.ReadAll(passwordFile)
	if err != nil {
		return err
	}

	if string(pwd) != password {
		return errors.New("password error")
	}
	return nil
}

func UserCreate(username string, password string) error {
	bucket := getBucket()
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
