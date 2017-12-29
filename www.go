/*
Copyright (C) 2017 by Ahmet Inan <inan@aicodix.de>

Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/


package main
import (
	"os"
	"path"
	"bytes"
	"bufio"
	"strings"
	"net/http"
	"io/ioutil"
	"gopkg.in/russross/blackfriday.v2"
)

func hostname() string { return "localhost" }
func assets() string { return "assets" }
func certificate() string { return "cer" }
func private_key() string { return "key" }

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://" + hostname() + r.URL.Path, http.StatusMovedPermanently)
}

func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

type markdownHandler struct {
	root http.FileSystem
	fsrv http.Handler
}

func Markdown(root http.FileSystem) http.Handler {
	return &markdownHandler{root, http.FileServer(root)}
}

func (f *markdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	// TODO: check if fs has index.md
	if strings.HasSuffix(upath, "/") {
		upath += "index.md"
	}
	if !strings.HasSuffix(upath, ".md") {
		f.fsrv.ServeHTTP(w, r)
		return
	}
	serveMarkdown(w, r, f.root, path.Clean(upath))
}

func serveMarkdown(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) {
	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	if d.IsDir() {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	title := d.Name()
	meta := ""
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		if strings.HasSuffix(scanner.Text(), ")") {
			tmp := strings.TrimSuffix(scanner.Text(), ")")
			if strings.HasPrefix(tmp, "[title]: # (") {
				title = strings.TrimPrefix(tmp, "[title]: # (")
			} else if strings.HasPrefix(tmp, "[meta]: # (") {
				meta += "<meta " + strings.TrimPrefix(tmp, "[meta]: # (") + " />\n"
			}
		}
	}
	output := `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>` + title + `</title>
` + meta + `</head>
<body>
` + string(blackfriday.Run(b)) + `
</body>
</html>`
	reader := bytes.NewReader([]byte(output))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, d.Name(), d.ModTime(), reader)
}

func main() {
	go http.ListenAndServe(":80", http.HandlerFunc(redirect))
	//go http.ListenAndServe(":80", http.RedirectHandler("https://" + hostname() + "/", http.StatusMovedPermanently))
	err := http.ListenAndServeTLS(":443", certificate(), private_key(), Markdown(http.Dir(assets())))
	if err != nil { panic(err) }
}

