package cement

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// usage
// err := Serve(":8080", "files")

func Serve(bind, keyPrefix string) error {
	bucket := getBucket()

	return http.ListenAndServe(bind, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Method:", r.Method)
		fmt.Println("URL:", r.URL)
		fmt.Println("Proto:", r.Proto)
		fmt.Println("Header:", r.Header)
		fmt.Println("Body:", r.Body)

		if r.Method == http.MethodOptions {
			w.Header().Set("DAV", "1, 2")
			w.Header().Set("MS-Author-Via", "DAV")
			w.Header().Set("Allow", "OPTIONS, GET, HEAD, DELETE, PUT, PROPFIND, PROPPATCH, COPY, MOVE, LOCK, UNLOCK")
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			isObjectExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isObjectExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			file, err := bucket.GetObject(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer file.Close()
			// get oss object meta
			meta, err := bucket.GetObjectDetailedMeta(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// set response header
			for k, v := range meta {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			// set response status code
			w.WriteHeader(http.StatusOK)
			// copy object to response writer
			_, err = io.Copy(w, file)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
		if r.Method == http.MethodHead {
			isObjectExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isObjectExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// get oss object meta
			meta, err := bucket.GetObjectDetailedMeta(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// set response header
			for k, v := range meta {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			// set response status code
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPut {
			cache_control := r.Header.Get("Cache-Control")
			content_disposition := r.Header.Get("Content-Disposition")
			content_encoding := r.Header.Get("Content-Encoding")
			content_language := r.Header.Get("Content-Language")
			content_type := r.Header.Get("Content-Type")
			expires, err := http.ParseTime(r.Header.Get("Expires"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// set oss object meta
			options := []oss.Option{
				oss.CacheControl(cache_control),
				oss.ContentDisposition(content_disposition),
				oss.ContentEncoding(content_encoding),
				oss.ContentLanguage(content_language),
				oss.ContentType(content_type),
				oss.Expires(expires),
			}

			// put object
			err = bucket.PutObject(keyPrefix+r.URL.Path, r.Body, options...)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodDelete {
			// if object not exist, return 404
			isObjectExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isObjectExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			err = bucket.DeleteObject(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "PROPFIND" {
			// if is a directory
			if r.URL.Path[len(r.URL.Path)-1] == '/' {
				isObjectExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if !isObjectExist {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// list objects
				marker := oss.Marker("")
				prefix := oss.Prefix(keyPrefix + r.URL.Path)
				var objects []oss.ObjectProperties
				// objects = append(objects, *ossProperties)
				for {
					lor, err := bucket.ListObjects(marker, prefix)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					objects = append(objects, lor.Objects...)
					if lor.IsTruncated {
						marker = oss.Marker(lor.NextMarker)
					} else {
						break
					}
				}

				// w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/xml; charset=UTF-8")

				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
				w.Write([]byte(`<D:multistatus xmlns:D="DAV:">`))
				for _, object := range objects {
					w.Write([]byte(`<D:response>`))
					w.Write([]byte(`<D:href>`))
					if object.Key != r.URL.Path {
						w.Write([]byte(r.URL.Path + object.Key[len(keyPrefix+r.URL.Path):]))
					} else {
						w.Write([]byte(r.URL.Path))
					}
					w.Write([]byte(`</D:href>`))
					w.Write([]byte(`<D:propstat>`))
					w.Write([]byte(`<D:prop>`))
					w.Write([]byte(`<D:getcontentlength>`))
					w.Write([]byte(strconv.FormatInt(object.Size, 10)))
					w.Write([]byte(`</D:getcontentlength>`))
					w.Write([]byte(`<D:getcontenttype>`))
					// if object is a directory, set content type to http.DetectContentType([]byte(""))
					if object.Key != r.URL.Path && object.Key[len(keyPrefix+r.URL.Path):] != "" {
						// if content type is empty, set it to application/octet-stream
						if http.DetectContentType([]byte(object.Key)) == "" {
							w.Write([]byte("application/octet-stream"))
						} else if http.DetectContentType([]byte(object.Key)) == "text/plain" {
							// get file extension
							fileExtension := filepath.Ext(object.Key)
							// if file extension is not empty, infer content type from file extension
							if fileExtension != "" {
								// get content type from file extension
								contentType := mime.TypeByExtension(fileExtension)
								// if content type is not empty, set it to content type
								if contentType != "" {
									w.Write([]byte(contentType))
								}
							}
						} else {
							w.Write([]byte(http.DetectContentType([]byte(object.Key))))
						}
					}
					w.Write([]byte(`</D:getcontenttype>`))
					w.Write([]byte(`<D:getlastmodified>`))
					w.Write([]byte(object.LastModified.Format(time.RFC1123)))
					w.Write([]byte(`</D:getlastmodified>`))
					w.Write([]byte(`<D:getetag>`))
					w.Write([]byte(object.ETag))
					w.Write([]byte(`</D:getetag>`))
					w.Write([]byte(`<D:supportedlock>`))
					w.Write([]byte(`<D:lockentry xmlns:D="DAV:">`))
					w.Write([]byte(`<D:lockscope><D:exclusive/></D:lockscope>`))
					w.Write([]byte(`<D:locktype><D:write/></D:locktype>`))
					w.Write([]byte(`</D:lockentry>`))
					w.Write([]byte(`</D:supportedlock>`))
					w.Write([]byte(`<D:resourcetype>`))
					if object.Key == r.URL.Path {
						w.Write([]byte(`<D:collection/>`))
					}
					w.Write([]byte(`</D:resourcetype>`))
					w.Write([]byte(`<D:displayname>`))
					if object.Key != r.URL.Path {
						w.Write([]byte(object.Key[len(keyPrefix+r.URL.Path):]))
					}
					w.Write([]byte(`</D:displayname>`))
					w.Write([]byte(`</D:prop>`))
					w.Write([]byte(`<D:status>HTTP/1.1 200 OK</D:status>`))
					w.Write([]byte(`</D:propstat>`))
					w.Write([]byte(`</D:response>`))
				}
				w.Write([]byte(`</D:multistatus>`))

				return
			} else {
				// if object not exist, return 404
				isObjectExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if !isObjectExist {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				// get oss object meta
				meta, err := bucket.GetObjectDetailedMeta(keyPrefix + r.URL.Path)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// set response header
				for k, v := range meta {
					for _, vv := range v {
						w.Header().Add(k, vv)
					}
				}

				// make a propfind response
				var propfindResponse = `<?xml version="1.0" encoding="UTF-8" ?>
<D:multistatus xmlns:D="DAV:">
<D:response>
<D:href>` + r.URL.Path + `</D:href>
<D:propstat>
<D:prop>
<D:resourcetype>
<D:collection/>
</D:resourcetype>
<D:getcontenttype>` + meta.Get("Content-Type") + `</D:getcontenttype>
<D:getcontentlength>` + meta.Get("Content-Length") + `</D:getcontentlength>
<D:creationdate>` + meta.Get("Last-Modified") + `</D:creationdate>
<D:getlastmodified>` + meta.Get("Last-Modified") + `</D:getlastmodified>
</D:prop>
<D:status>HTTP/1.1 200 OK</D:status>
</D:propstat>
</D:response>
</D:multistatus>
`

				// set response status code
				w.WriteHeader(http.StatusOK)

				// write response
				_, err = w.Write([]byte(propfindResponse))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}
