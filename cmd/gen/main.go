package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"text/template"
	"time"
        "log"

	"github.com/otiai10/copy"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/renderer"
)

type post struct {
    path          string
    index         uint32
    last_modified int64

    title    string
    subtitle string
    author   string
    content  string
}

// utility functions
func read_into_string(path string) string {
    file, err := os.Open(path)
    if err != nil {
        panic(err)
    }

    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)

    var text string
    for scanner.Scan() {
        // beware of my shit code
        text = text + scanner.Text() + "\n"
    }

    file.Close()
    return text
}

func min(a int, b int) int {
    if (a < b) {
        return a
    }
    return b
}

// other functions
func find_post_files() []post {
    posts :=      []post{}

    file, err := os.Open("posts")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    fileList,_ := file.Readdir(0)
    for _, files := range fileList {
        if(files.IsDir()) {
            continue
        }

        new_post := post{}

        var _, err = fmt.Sscanf(files.Name(), "p%d", &new_post.index)
        if(err != nil) {
            continue
        }

        new_post.path = "posts/" + files.Name()
        new_post.last_modified = files.ModTime().Unix()

        posts = append(posts, new_post)
    }

    return posts
}

func get_data_string(name string, data map[string]interface {}, default_value string) (string, bool) {
    var val string
    var ok bool
    if x, found := data[name]; found {
        if val, ok = x.(string); !ok {
            panic(name + " wasn't a string. FIX THIS!")
        }
        return val, found
    } else {
        return default_value, found
    }
}

func load_post_files(posts []post) {
    markdown := goldmark.New(
        goldmark.WithExtensions(
            meta.New(meta.WithStoresInDocument()),
        ),
        goldmark.WithRendererOptions(
            renderer.WithNodeRenderers(
            ),
        ),
    )

    for i := 0; i < len(posts); i++ {
        var buf bytes.Buffer
        var ok  bool

        p := posts[i]
        source := read_into_string(p.path)

        context := parser.NewContext()
        err := markdown.Convert([]byte(source), &buf, parser.WithContext(context))
        if err != nil {
            panic(err)
        }

        metaData := meta.Get(context)

        p.content = buf.String()
        if p.title,    ok = get_data_string("Title", metaData, "NO TITLE"); !ok {
            log.Print("Title not set in " + p.path)
        }
        if p.author,   ok = get_data_string("Author", metaData, "NO AUTHOR"); !ok {
            log.Print("Author not set in " + p.path)
        }
        if p.subtitle, ok = get_data_string("Subtitle", metaData, "NO SUBTITLE"); !ok {
            log.Print("Subtitle not set in " + p.path)
        }
        posts[i] = p
    }

    // reverse chronological sorting, ie. highest number first
    sort.Slice(posts,
        func(i, j int) bool {
            return posts[i].index > posts[j].index
        })
}

type index_generator_info struct {
    Content string
}

func generate_index_site_from_posts(posts []post) {
    ut, err := template.New("site").Parse(read_into_string("templates/index.html"))
    if err != nil {
        panic(err)
    }

    buf := new(bytes.Buffer)
    site := index_generator_info{}

    for i := 0; i < min(len(posts), 3); i++ {
        site.Content += "<div class=\"bg-dark text-white flex-fix-news text-center p-2\">"
        site.Content += posts[i].content
        site.Content += "</div>"
    }

    ut.Execute(buf, site)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile("generated/index.html", buf.Bytes(), 0777)
    if err != nil {
        panic(err)
    }
}

type post_generator_info struct {
    Content string
    Date    string
    Title   string
    Author  string
}

func generate_post(post post) {
    ut, err := template.New("site").Parse(read_into_string("templates/page.html"))
    if err != nil {
        panic(err)
    }

    buf := new(bytes.Buffer)
    site := post_generator_info{}

    site.Content = post.content
    site.Author = post.author
    site.Date = time.Unix(post.last_modified, 0).Format("02/01/2006")
    site.Title = post.title

    ut.Execute(buf, site)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile("generated/post/"+strconv.Itoa(int(post.index))+".html",
                            buf.Bytes(), 0777)
    if err != nil {
        panic(err)
    }
}

func prepare_generated_folder_tree() {
    os.RemoveAll("generated")

    dirs := []string{
        "generated",
        "generated/post",
    }

    for _, folder := range dirs {
        err := os.Mkdir(folder, 0755)
        if err != nil {
            panic(err)
        }
    }

    err := copy.Copy("files", "generated")
    if err != nil {
        panic(err)
    }
}

func main() {
    prepare_generated_folder_tree()
    posts := find_post_files()
    load_post_files(posts)
    generate_index_site_from_posts(posts)
    for _, post := range posts {
        generate_post(post)
    }
}

