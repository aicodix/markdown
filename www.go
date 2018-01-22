/*
Markdown web server

Copyright 2017 Ahmet Inan <inan@aicodix.de>
*/

package main
import (
	"io"
	"os"
	"path"
	"time"
	"bytes"
	"bufio"
	"errors"
	"strings"
	"net/http"
	"io/ioutil"
	"gopkg.in/russross/blackfriday.v2"
)

func hostname() string { return "localhost" }
func assets() string { return "assets" }
func certificate() string { return "cer" }
func private_key() string { return "key" }
func template_html(string) string { return "/template.html" }

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

func readFileOrFail(w http.ResponseWriter, fs http.FileSystem, name string) ([]byte, string, time.Time, error) {
	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return nil, "", time.Time{}, err
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return nil, "", time.Time{}, err
	}
	if d.IsDir() {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return nil, "", time.Time{}, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return nil, "", time.Time{}, err
	}
	return b, d.Name(), d.ModTime(), nil
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
	if strings.HasSuffix(upath, "/") {
		index := upath + "index.md"
		file, err := f.root.Open(index)
		if err == nil {
			file.Close()
			upath = index
		}
	}
	if !strings.HasSuffix(upath, ".md") {
		f.fsrv.ServeHTTP(w, r)
		return
	}
	serveMarkdown(w, r, f.root, path.Clean(upath))
}

func parseMetadata(r io.Reader) (string, string) {
	title := ""
	head := ""
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		tmp := strings.SplitN(scanner.Text(), "]: # (", 2)
		if len(tmp) != 2 { return title, head }

		if !strings.HasPrefix(tmp[0], "[") { return title, head }
		key := strings.TrimPrefix(tmp[0], "[")

		if !strings.HasSuffix(tmp[1], ")") { return title, head }
		value := strings.TrimSuffix(tmp[1], ")")

		switch key {
		case "title":
			title = value
		case "head":
			head += value + "\n"
		default:
			return title, head
		}
	}
	return title, head
}

func fillInTheBlanks(tmpl, head, body string) (string, error) {
	s := strings.Split(tmpl, "<!--here-->")
	if len(s) != 3 { return "", errors.New("Template error") }
	return s[0] + head + s[1] + body + s[2], nil
}

func serveMarkdown(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) {
	md_bytes, md_name, md_modtime, err := readFileOrFail(w, fs, name)
	if err != nil { return }

	tl_bytes, _, tl_modtime, err := readFileOrFail(w, fs, template_html(name))
	if err != nil { return }

	tmpl := string(tl_bytes)
	modtime := md_modtime
	if modtime.Before(tl_modtime) { modtime = tl_modtime }

	title, head := parseMetadata(bytes.NewReader(md_bytes))
	if title == "" { title = md_name }
	head += "<title>" + title + "</title>\n"
	body := string(blackfriday.Run(md_bytes))
	output, err := fillInTheBlanks(tmpl, head, body)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	reader := bytes.NewReader([]byte(output))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, md_name, modtime, reader)
}

func main() {
	go http.ListenAndServe(":80", http.HandlerFunc(redirect))
	//go http.ListenAndServe(":80", http.RedirectHandler("https://" + hostname() + "/", http.StatusMovedPermanently))
	err := http.ListenAndServeTLS(":443", certificate(), private_key(), Markdown(http.Dir(assets())))
	if err != nil { panic(err) }
}

