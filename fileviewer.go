package cement

import (
	"io/ioutil"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "" {
		path = "."
	}
	path = "fileviewer/" + path
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		index := path + "/index.html"
		if _, err := os.Stat(index); err == nil {
			http.ServeFile(w, r, index)
		} else {
			http.Error(w, "can't read directory", http.StatusBadRequest)
			return
		}
	} else {
		data, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}
}

func Serve() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
