package cement

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func getCaptcha(timestamp string) string {
	url := "https://webservice.forclass.net/Account/GetRegisterCaptcha?stamp=" + timestamp
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(body)
}

func checkCaptcha(captcha string, timestamp string) bool {
	url := "https://webservice.forclass.net/Account/ValidateRegisterCode?stamp=" + timestamp + "&code=" + captcha
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	// json parse body
	var data map[string]interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return false
	}
	retcode := data["retcode"].(float64)

	if retcode == 0 {
		return true
	} else {
		return false
	}
}
