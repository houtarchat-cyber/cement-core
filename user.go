package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
)

func userLogin(username string, password string) string {
	bucket := getBucket()
	username = base64.StdEncoding.EncodeToString([]byte(username))
	password = fmt.Sprintf("%x", md5.Sum([]byte(password)))

	isExist, err := bucket.IsObjectExist("users/" + username)
	if err != nil {
		return err.Error()
	}
	if !isExist {
		return "user not exist"
	}
	pwdfile, err := bucket.GetObject("users/" + username)
	if err != nil {
		return err.Error()
	}

	defer pwdfile.Close()

	pwd, err := ioutil.ReadAll(pwdfile)
	if err != nil {
		return err.Error()
	}

	if string(pwd) != password {
		return "password error"
	}
	return "login success"
}

func userRegister(username string, password string) string {
	bucket := getBucket()
	username = base64.StdEncoding.EncodeToString([]byte(username))
	password = fmt.Sprintf("%x", md5.Sum([]byte(password)))

	isExist, err := bucket.IsObjectExist("users/" + username)
	if err != nil {
		return err.Error()
	}
	if isExist {
		return "user already exist"
	}

	err = bucket.PutObject("users/"+username, strings.NewReader(password))
	if err != nil {
		return err.Error()
	}
	return "register success"
}
