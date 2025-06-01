package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

// PostMetadata represents the metadata for a blog post
type PostMetadata struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Description string    `yaml:"description"`
	Tags        []string  `yaml:"tags"`
}

// Post represents a complete blog post
type Post struct {
	Metadata PostMetadata
	Content  template.HTML
	Slug     string
}

// PageData represents data passed to templates
type PageData struct {
	Title   string
	Content template.HTML
	Posts   []Post
	Post    *Post
}

var (
	templates *template.Template
	md        goldmark.Markdown
)

func init() {
	// Initialize goldmark with extensions
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

func main() {
	// Parse templates
	var err error
	templates, err = template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/posts/", postsHandler)
	http.HandleFunc("/post/", postHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Read content from title-page directory
	content, err := readMarkdownFile("title-page/index.md")
	if err != nil {
		log.Printf("Error reading title page: %v", err)
		content = template.HTML("<p>Welcome to my blog!</p>")
	}

	data := PageData{
		Title:   "My Personal Website",
		Content: content,
	}

	renderTemplate(w, "home.html", data)
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := loadAllPosts()
	if err != nil {
		http.Error(w, "Error loading posts", http.StatusInternalServerError)
		log.Printf("Error loading posts: %v", err)
		return
	}

	data := PageData{
		Title: "Blog Posts",
		Posts: posts,
	}

	renderTemplate(w, "posts.html", data)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Extract slug from URL
	slug := strings.TrimPrefix(r.URL.Path, "/post/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	post, err := loadPost(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title: post.Metadata.Title,
		Post:  post,
	}

	renderTemplate(w, "post.html", data)
}

func loadAllPosts() ([]Post, error) {
	var posts []Post

	blogsDir := "blogs"
	entries, err := os.ReadDir(blogsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		post, err := loadPost(entry.Name())
		if err != nil {
			log.Printf("Error loading post %s: %v", entry.Name(), err)
			continue
		}

		posts = append(posts, *post)
	}

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Metadata.Date.After(posts[j].Metadata.Date)
	})

	return posts, nil
}

func loadPost(slug string) (*Post, error) {
	postDir := filepath.Join("blogs", slug)

	// Read metadata
	metadataPath := filepath.Join(postDir, "metadata.yaml")
	metadata, err := readMetadata(metadataPath)
	if err != nil {
		return nil, err
	}

	// Read content
	contentPath := filepath.Join(postDir, "index.md")
	content, err := readMarkdownFile(contentPath)
	if err != nil {
		return nil, err
	}

	return &Post{
		Metadata: *metadata,
		Content:  content,
		Slug:     slug,
	}, nil
}

func readMetadata(path string) (*PostMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata PostMetadata
	err = yaml.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func readMarkdownFile(path string) (template.HTML, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}
