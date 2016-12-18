package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gernest/front"
)

// Post is a single post summary for the menu.
type Post struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Summary string `json:"summary,omitifempty"`
	Body    string `json:"body, omitifempty"`
	Date    string `json:"date"`
}

// Posts implements sort.Interface for a slice of Post objects.
type Posts []*Post

func (p Posts) Len() int { return len(p) }
func (p Posts) Less(i, j int) bool {
	iDate, _ := time.Parse("2006-01-02", p[i].Date)
	jDate, _ := time.Parse("2006-01-02", p[j].Date)

	return iDate.Unix() < jDate.Unix()
}
func (p Posts) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

var posts Posts

func init() {
	err := filepath.Walk("./blog/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fin, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fin.Close()

		content, err := ioutil.ReadAll(fin)
		if err != nil {
			// handle error
		}

		m := front.NewMatter()
		m.Handle("---", front.YAMLHandler)
		front, _, err := m.Parse(bytes.NewReader(content))
		if err != nil {
			return err
		}

		sp := strings.Split(string(content), "\n")
		sp = sp[4:]
		data := strings.Join(sp, "\n")

		p := &Post{
			Title: front["title"].(string),
			Date:  front["date"].(string),
			Link:  strings.Split(path, ".")[0],
			Body:  data,
		}

		posts = append(posts, p)

		return nil
	})

	if err != nil {
		panic(err)
	}

	sort.Sort(sort.Reverse(posts))
}

func main() {
	http.HandleFunc("/api/blog/posts", writeBlogPosts)
	http.HandleFunc("/api/blog/post", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := q.Get("name")

		if name == "" {
			goto fail
		}

		for _, p := range posts {
			if strings.HasSuffix(p.Link, name) {
				json.NewEncoder(w).Encode(p)
				return
			}
		}

	fail:
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	http.Handle("/dist/", http.FileServer(http.Dir("./frontend/static/")))
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", writeIndexHTML)

	log.Fatal(http.ListenAndServe(":9090", nil))
}

func writeBlogPosts(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(posts)
}

func writeIndexHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/static/dist/index.html")
}