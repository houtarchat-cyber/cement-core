package cement

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dreamacro/clash/component/mmdb"
	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub/executor"
	"github.com/Dreamacro/clash/log"
)

func GetClashConfig(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func Proxy(conf []byte) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	constant.SetHomeDir(currentDir)

	err = initMMDB()
	if err != nil {
		return err
	}
	cfg, err := executor.ParseWithBytes(conf)
	if err != nil {
		return err
	}

	executor.ApplyConfig(cfg, true)

	// wait for signal (keep process alive)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	return nil
}

func initMMDB() error {
	if _, err := os.Stat(constant.Path.MMDB()); os.IsNotExist(err) {
		log.Infoln("Can't find MMDB, start download")
		if err := downloadMMDB(constant.Path.MMDB()); err != nil {
			return fmt.Errorf("can't download MMDB: %s", err.Error())
		}
	}

	if !mmdb.Verify() {
		log.Warnln("MMDB invalid, remove and download")
		if err := os.Remove(constant.Path.MMDB()); err != nil {
			return fmt.Errorf("can't remove invalid MMDB: %s", err.Error())
		}

		if err := downloadMMDB(constant.Path.MMDB()); err != nil {
			return fmt.Errorf("can't download MMDB: %s", err.Error())
		}
	}

	return nil
}

func downloadMMDB(path string) (err error) {
	resp, err := http.Get("https://cdn.jsdelivr.net/gh/Dreamacro/maxmind-geoip@release/Country.mmdb")
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser, err *error) {
		*err = Body.Close()
	}(resp.Body, &err)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func(f *os.File, err *error) {
		*err = f.Close()
	}(f, &err)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)

	return err
}
