package cement

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetCaptcha(timestamp string) string {
	url := "https://webservice.forclass.net/Account/GetRegisterCaptcha?stamp=" + timestamp
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer func(Body io.ReadCloser, err *error) {
		*err = Body.Close()
	}(resp.Body, &err)
	if err != nil {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(body)
}

func CheckCaptcha(captcha string, timestamp string) bool {
	url := "https://webservice.forclass.net/Account/ValidateRegisterCode?stamp=" + timestamp + "&code=" + captcha
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer func(Body io.ReadCloser, err *error) {
		*err = Body.Close()
	}(resp.Body, &err)
	if err != nil {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	// json parse body
	var data map[string]interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return false
	}
	//goland:noinspection SpellCheckingInspection
	retCode := data["retcode"].(float64)

	if retCode == 0 {
		return true
	} else {
		return false
	}
}

func GetTimestamp() string {
	// get timestamp
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d", timestamp)
}
