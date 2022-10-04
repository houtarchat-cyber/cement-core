package cement

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func Serve(bind, keyPrefix string) error {
	bucket := GetBucket()
	keyPrefix = "files/" + base64.StdEncoding.EncodeToString([]byte(keyPrefix))

	return http.ListenAndServe(bind, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Method:", r.Method)
		fmt.Println("URL:", r.URL)
		fmt.Println("Proto:", r.Proto)
		fmt.Println("Header:", r.Header)
		fmt.Println("Body:", r.Body)

		if r.Method == http.MethodOptions {
			w.Header().Set("DAV", "1")
			w.Header().Set("MS-Author-Via", "DAV")
			w.Header().Set("Allow", "OPTIONS, PROPFIND, PROPPATCH, MKCOL, GET, HEAD, POST, DELETE, PUT, COPY, MOVE, LOCK, UNLOCK")
			w.Header().Set("Content-Length", "0")
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

				w.WriteHeader(http.StatusMultiStatus)
				w.Header().Set("Content-Type", "application/xml; charset=UTF-8")

				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
				w.Write([]byte(`<D:multistatus xmlns:D="DAV:">`))
				for _, object := range objects {
					w.Write([]byte(`<D:response>`))
					w.Write([]byte(`<D:href>`))
					if object.Key != r.URL.Path {
						w.Write([]byte(`http://` + r.Host + r.URL.Path + object.Key[len(keyPrefix+r.URL.Path):]))
						// w.Write([]byte(r.URL.Path + object.Key[len(keyPrefix+r.URL.Path):]))
					} else {
						w.Write([]byte(`http://` + r.Host + r.URL.Path))
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
<D:href>` + `http://` + r.Host + r.URL.Path + `</D:href>
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
				w.WriteHeader(http.StatusMultiStatus)

				// write response
				_, err = w.Write([]byte(propfindResponse))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}
		}
		if r.Method == "PROPPATCH" {
			// return 500 because PROPPATCH is not supported
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.Method == "MKCOL" {
			if r.URL.Path == "/" {
				w.WriteHeader(http.StatusConflict)
				return
			}
			// check if parent dir exist
			// for example, if request path is /a/b/c, we need to check if /a/b exist
			parentDir := path.Dir(r.URL.Path)
			if parentDir == "." {
				parentDir = "/"
			}
			isParentDirExist, err := bucket.IsObjectExist(keyPrefix + parentDir)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isParentDirExist {
				w.WriteHeader(http.StatusConflict)
				return
			}

			// if request content length is 0, create a dir
			if r.ContentLength == 0 {
				err = bucket.PutObject(keyPrefix+r.URL.Path, strings.NewReader(""))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusCreated)
				return
			} else {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
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
			if strings.HasSuffix(r.URL.Path, "/") {
				// get all objects under the dir
				objects, err := bucket.ListObjects(oss.Prefix(keyPrefix + r.URL.Path))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// get html from mount.html
				htmlB, err := os.ReadFile("mount.html")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				html := string(htmlB)
				html += `<script>start("` + r.URL.Path + `");</script>`
				// if not root dir, add parent dir link
				if r.URL.Path != "/" {
					html += `<script>onHasParentDirectory();</script>`
				}
				for _, object := range objects.Objects {
					if object.Key == keyPrefix+r.URL.Path {
						continue
					}
					name := object.Key[len(keyPrefix)+1:]
					url := object.Key[len(keyPrefix)+1:]
					isDir := "0"
					if strings.HasSuffix(object.Key, "/") {
						isDir = "1"
						name = name[:len(name)-1]
						url = url[:len(url)-1]
					}
					size := object.Size
					// human readable size
					var sizeStr string
					if size < 1024 {
						sizeStr = fmt.Sprintf("%d B", size)
					} else if size < 1024*1024 {
						sizeStr = fmt.Sprintf("%.2f KB", float64(size)/1024)
					} else if size < 1024*1024*1024 {
						sizeStr = fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
					} else {
						sizeStr = fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
					}
					lastModified := object.LastModified
					// human readable last modified
					var lastModifiedStr string
					// if time.Now().Sub(lastModified) < 24*time.Hour {
					if time.Since(lastModified) < 24*time.Hour {
						lastModifiedStr = lastModified.In(time.Local).Format("15:04")
					} else {
						lastModifiedStr = lastModified.In(time.Local).Format("2006-01-02")
					}
					html += `<script>addRow("` + name + `","` + url + `",` + isDir + `,` +
						strconv.FormatInt(size, 10) + `,"` + sizeStr + `",` +
						strconv.FormatInt(lastModified.Unix(), 10) + `,"` +
						lastModifiedStr + `");</script>`
				}
				// set response header
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("Content-Length", strconv.Itoa(len(html)))
				// write response
				_, err = w.Write([]byte(html))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			} else {
				// if is a file, return the file
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
		if r.Method == http.MethodPost {
			// get file from request
			file, header, err := r.FormFile("file")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer file.Close()
			// get file name
			name := header.Filename
			// get file size
			size := header.Size
			// get file mime type
			mimeType := header.Header.Get("Content-Type")
			// get file content
			content, err := io.ReadAll(file)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// upload file to oss
			err = bucket.PutObject(keyPrefix+r.URL.Path+name, bytes.NewReader(content))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// set response header
			w.Header().Set("Content-Type", "application/json")
			// write response
			_, err = w.Write([]byte(`{"name":"` + name + `","size":` + strconv.FormatInt(size, 10) +
				`,"type":"` + mimeType + `"}`))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
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
			// if is a directory, delete all objects in the directory
			if strings.HasSuffix(r.URL.Path, "/") {
				// list all objects in the directory
				marker := oss.Marker("")
				result := `<?xml version="1.0" encoding="utf-8" ?> 
				<d:multistatus xmlns:d="DAV:">`
				hasErr := false
				for {
					lor, err := bucket.ListObjects(oss.Prefix(keyPrefix+r.URL.Path), marker)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					for _, object := range lor.Objects {
						// delete object
						err = bucket.DeleteObject(object.Key)
						if err != nil {
							hasErr = true
							result += `
							<d:response> 
							  <d:href>` + r.URL.Path + `</d:href> 
							  <d:status>HTTP/1.1 500 Internal Server Error</d:status> 
							  <d:error></d:error>
							</d:response> `
							continue
						}
					}
					if lor.IsTruncated {
						marker = oss.Marker(lor.NextMarker)
					} else {
						break
					}
				}
				result += `</d:multistatus>`
				if hasErr {
					// set response status code
					w.WriteHeader(http.StatusMultiStatus)
					// set response header
					w.Header().Set("Content-Type", `application/xml; charset="utf-8"`)
					w.Header().Set("Content-Length", strconv.Itoa(len(result)))
					// write response
					_, err = w.Write([]byte(result))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
				return
			} else {
				err = bucket.DeleteObject(keyPrefix + r.URL.Path)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		if r.Method == http.MethodPut {
			if strings.HasSuffix(r.URL.Path, "/") {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

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

			parentDir := path.Dir(r.URL.Path)
			if parentDir == "." {
				parentDir = "/"
			}
			isParentDirExist, err := bucket.IsObjectExist(keyPrefix + parentDir)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isParentDirExist {
				w.WriteHeader(http.StatusConflict)
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
		if r.Method == "COPY" {
			// request like:
			// COPY /~fielding/index.html HTTP/1.1
			// Host: www.example.com
			// Destination: http://www.example.com/users/f/fielding/index.html

			isSourceExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isSourceExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// check if destination object exist
			destination := r.Header.Get("Destination")
			if destination == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if !strings.HasSuffix(r.URL.Path, "/") {
				// COPY for Non-collection Resources
				// copy object
				// if no overwrite
				if r.Header.Get("Overwrite") == "F" {
					isDestExist, err := bucket.IsObjectExist(keyPrefix + destination)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if isDestExist {
						w.WriteHeader(http.StatusPreconditionFailed)
						return
					}
				}
				_, err = bucket.CopyObject(keyPrefix+r.URL.Path, keyPrefix+destination)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusCreated)
				return

			} else {
				// COPY for Collection Resources
				// list all objects in the directory
				marker := oss.Marker("")
				result := `<?xml version="1.0" encoding="utf-8" ?><d:multistatus xmlns:d="DAV:">`
				hasErr := false
				for {
					lor, err := bucket.ListObjects(oss.Prefix(keyPrefix+r.URL.Path), marker)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					for _, object := range lor.Objects {
						// copy object
						// if no overwrite
						if r.Header.Get("Overwrite") == "F" {
							isDestExist, err := bucket.IsObjectExist(keyPrefix + destination + strings.TrimPrefix(object.Key, keyPrefix+r.URL.Path))
							if err != nil {
								hasErr = true
								result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 500 Internal Server Error</d:status> 
											  <d:error></d:error>
											</d:response> `
								continue
							}
							if isDestExist {
								hasErr = true
								result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 412 Precondition Failed</d:status> 
											  <d:error></d:error>
											</d:response> `
								continue
							}
						}
						_, err = bucket.CopyObject(object.Key, keyPrefix+destination+strings.TrimPrefix(object.Key, keyPrefix+r.URL.Path))
						if err != nil {
							hasErr = true
							result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 500 Internal Server Error</d:status> 
											  <d:error></d:error>
											</d:response> `
							continue
						}
					}
					if lor.IsTruncated {
						marker = oss.Marker(lor.NextMarker)
					} else {
						break
					}
				}
				result += `</d:multistatus>`
				if hasErr {
					// set response status code
					w.WriteHeader(http.StatusMultiStatus)
					// set response header
					w.Header().Set("Content-Type", `application/xml; charset="utf-8"`)
					w.Header().Set("Content-Length", strconv.Itoa(len(result)))
					// write response
					_, err = w.Write([]byte(result))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
				return
			}
		}
		if r.Method == "MOVE" {
			// request like:
			// MOVE /~fielding/index.html HTTP/1.1
			// Host: www.example.com
			// Destination: http://www.example.com/users/f/fielding/index.html

			isSourceExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isSourceExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// check if destination object exist
			destination := r.Header.Get("Destination")
			if destination == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if !strings.HasSuffix(r.URL.Path, "/") {
				// COPY for Non-collection Resources
				// move object
				// if no overwrite
				if r.Header.Get("Overwrite") == "F" {
					isDestExist, err := bucket.IsObjectExist(keyPrefix + destination)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if isDestExist {
						w.WriteHeader(http.StatusPreconditionFailed)
						return
					}
				}
				_, err = bucket.CopyObject(keyPrefix+r.URL.Path, keyPrefix+destination)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// delete source object
				err = bucket.DeleteObject(keyPrefix + r.URL.Path)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusCreated)
				return

			} else {
				// MOVE for Collection Resources
				// list all objects in the directory
				marker := oss.Marker("")
				result := `<?xml version="1.0" encoding="utf-8" ?><d:multistatus xmlns:d="DAV:">`
				hasErr := false
				for {
					lor, err := bucket.ListObjects(oss.Prefix(keyPrefix+r.URL.Path), marker)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					for _, object := range lor.Objects {
						// move object
						// if no overwrite
						if r.Header.Get("Overwrite") == "F" {
							isDestExist, err := bucket.IsObjectExist(keyPrefix + destination + strings.TrimPrefix(object.Key, keyPrefix+r.URL.Path))
							if err != nil {
								hasErr = true
								result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 500 Internal Server Error</d:status> 
											  <d:error></d:error>
											</d:response> `
								continue
							}
							if isDestExist {
								hasErr = true
								result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 412 Precondition Failed</d:status> 
											  <d:error></d:error>
											</d:response> `
								continue
							}
						}
						_, err = bucket.CopyObject(object.Key, keyPrefix+destination+strings.TrimPrefix(object.Key, keyPrefix+r.URL.Path))
						if err != nil {
							hasErr = true
							result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 500 Internal Server Error</d:status> 
											  <d:error></d:error>
											</d:response> `
							continue
						}
						// delete source object
						err = bucket.DeleteObject(keyPrefix + r.URL.Path)
						if err != nil {
							hasErr = true
							result += `
											<d:response> 
											  <d:href>` + r.URL.Path + `</d:href> 
											  <d:status>HTTP/1.1 500 Internal Server Error</d:status> 
											  <d:error></d:error>
											</d:response> `
							continue
						}
					}
					if lor.IsTruncated {
						marker = oss.Marker(lor.NextMarker)
					} else {
						break
					}
				}
				result += `</d:multistatus>`
				if hasErr {
					// set response status code
					w.WriteHeader(http.StatusMultiStatus)
					// set response header
					w.Header().Set("Content-Type", `application/xml; charset="utf-8"`)
					w.Header().Set("Content-Length", strconv.Itoa(len(result)))
					// write response
					_, err = w.Write([]byte(result))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
				return
			}
		}
		if r.Method == "LOCK" {
			// LOCK
			// check if the object exists
			isExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// return 500 because we don't support lock
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.Method == "UNLOCK" {
			// UNLOCK
			// check if the object exists
			isExist, err := bucket.IsObjectExist(keyPrefix + r.URL.Path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !isExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// return 500 because we don't support lock
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}
