package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/houtarchat-cyber/cement-core"
)

//export UserLogin
func UserLogin(username, password string) error {
	return cement.UserLogin(username, password)
}

//export UserCreate
func UserCreate(username, password string) error {
	return cement.UserCreate(username, password)
}

//export GetTimestamp
func GetTimestamp() string {
	return cement.GetTimestamp()
}

//export GetCaptcha
func GetCaptcha(timestamp string) string {
	return cement.GetCaptcha(timestamp)
}

//export CheckCaptcha
func CheckCaptcha(captcha, timestamp string) bool {
	return cement.CheckCaptcha(captcha, timestamp)
}

//export Serve
func Serve(bind, keyPrefix string) error {
	return cement.Serve(bind, keyPrefix)
}

//export ChannelCreate
func ChannelCreate(channelName string) error {
	return cement.ChannelCreate(channelName)
}

//export ChannelSend
func ChannelSend(channelName, username, message string) error {
	return cement.ChannelSend(channelName, username, message)
}

//export ChannelReceive
func ChannelReceive(channelName string) ([]string, error) {
	_, jsons, err := cement.ChannelReceive(channelName)
	return jsons, err
}

//export Proxy
func Proxy(url string) error {
	return cement.Proxy(cement.GetClashConfig(url))
}

func main() {
	// 1. ./cement proxy <url>
	if len(os.Args) == 3 && os.Args[1] == "proxy" {
		err := Proxy(os.Args[2])
		if err != nil {
			panic(err)
		}
		return
	}
	// 2. ./cement serve <bind> <keyPrefix>
	if len(os.Args) == 4 && os.Args[1] == "serve" {
		err := Serve(os.Args[2], os.Args[3])
		if err != nil {
			panic(err)
		}
		return
	}
	// 3. ./cement channel create <channelName>
	if len(os.Args) == 4 && os.Args[1] == "channel" && os.Args[2] == "create" {
		err := ChannelCreate(os.Args[3])
		if err != nil {
			panic(err)
		}
		return
	}
	// 4. ./cement channel send <channelName> <username> <message>
	if len(os.Args) == 6 && os.Args[1] == "channel" && os.Args[2] == "send" {
		err := ChannelSend(os.Args[3], os.Args[4], os.Args[5])
		if err != nil {
			panic(err)
		}
		return
	}
	// 5. ./cement channel receive <channelName>
	if len(os.Args) == 4 && os.Args[1] == "channel" && os.Args[2] == "receive" {
		jsons, _, err := cement.ChannelReceive(os.Args[3])
		if err != nil {
			panic(err)
		}
		for _, json := range jsons {
			fmt.Println(json.Username+"于"+json.Time+"发送消息：", json.Message)
		}
		return
	}
	// 6. ./cement user login <username> <password>
	if len(os.Args) == 5 && os.Args[1] == "user" && os.Args[2] == "login" {
		err := UserLogin(os.Args[3], os.Args[4])
		if err != nil {
			panic(err)
		}
		return
	}
	// 7. ./cement user create <username> <password>
	if len(os.Args) == 5 && os.Args[1] == "user" && os.Args[2] == "create" {
		timestamp := GetTimestamp()
		captcha := GetCaptcha(timestamp)
		decoded, err := base64.StdEncoding.DecodeString(captcha)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("captcha.jpeg", decoded, 0644)
		if err != nil {
			panic(err)
		}

		if os.PathSeparator == '/' {
			_, err := os.Stat("/usr/bin/xdg-open")
			if err == nil {
				_, err = os.StartProcess("/usr/bin/xdg-open", []string{"xdg-open", "captcha.jpeg"}, &os.ProcAttr{})
				if err != nil {
					panic(err)
				}
			} else {
				_, err = os.Stat("/usr/bin/open")
				if err == nil {
					_, err = os.StartProcess("/usr/bin/open", []string{"open", "captcha.jpeg"}, &os.ProcAttr{})
					if err != nil {
						panic(err)
					}
				} else {
					fmt.Println("请手动打开captcha.jpeg")
				}
				if err != nil {
					panic(err)
				}
			}
		} else if os.PathSeparator == '\\' {
			_, err := os.Stat("C:\\Program Files\\Windows Photo Viewer\\PhotoViewer.dll")
			if err == nil {
				_, err = os.StartProcess("C:\\Program Files\\Windows Photo Viewer\\PhotoViewer.dll", []string{"PhotoViewer.dll", "captcha.jpeg"}, &os.ProcAttr{})
				if err != nil {
					panic(err)
				}
			} else {
				_, err = os.StartProcess("cmd", []string{"cmd", "/c", "start", "captcha.jpeg"}, &os.ProcAttr{})
				if err != nil {
					panic(err)
				}
			}
			if err != nil {
				panic(err)
			}
		}

		// check captcha
		var input string
		fmt.Print("请输入验证码：")
		fmt.Scanln(&input)
		if !CheckCaptcha(input, timestamp) {
			panic("验证码错误")
		}

		err = UserCreate(os.Args[3], os.Args[4])
		if err != nil {
			panic(err)
		}
		return
	}
	fmt.Println("Usage:")
	fmt.Println("  ./cement proxy <url>")
	fmt.Println("  ./cement serve <bind> <keyPrefix>")
	fmt.Println("  ./cement channel create <channelName>")
	fmt.Println("  ./cement channel send <channelName> <username> <message>")
	fmt.Println("  ./cement channel receive <channelName>")
	fmt.Println("  ./cement user login <username> <password>")
	fmt.Println("  ./cement user create <username> <password>")
}
