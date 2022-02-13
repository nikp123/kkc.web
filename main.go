package main

import (
    "github.com/gomarkdown/markdown"
    "os"
    "text/template"
    "fmt"
    "bufio"
    "bytes"
    "sort"
    "io/ioutil"
)

type post struct {
    path          string
    index         uint32
    last_modified int64

    title    string
    subtitle string
    author   string
    Content  string
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

func load_post_files(posts []post) {
    for i := 0; i < len(posts); i++ {
        p := posts[i]
        markdown_content := read_into_string(p.path)
        fmt.Printf("%s\n",markdown_content)
        p.Content = string(markdown.ToHTML([]byte(markdown_content), nil, nil))
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
        site.Content += posts[i].Content
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

func main() {
    posts := find_post_files()
    load_post_files(posts)
    generate_index_site_from_posts(posts)
}

