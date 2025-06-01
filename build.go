package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	// "strings"
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
	// Clean output directory
	os.RemoveAll("public")
	os.MkdirAll("public", 0755)
	os.MkdirAll("public/post", 0755)

	// Copy static files
	copyDir("static", "public/static")

	// Parse templates
	var err error
	templates, err = template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// Generate pages
	generateHomePage()
	posts := generatePostPages()
	generatePostsListPage(posts)

	// Copy blog images
	copyBlogImages()

	fmt.Println("✅ Site built successfully in ./public")
}

func generateHomePage() {
	content, err := readMarkdownFile("title-page/index.md")
	if err != nil {
		log.Printf("Error reading title page: %v", err)
		content = template.HTML("<p>Welcome to my blog!</p>")
	}

	data := PageData{
		Title:   "My Personal Website",
		Content: content,
	}

	renderToFile("public/index.html", "home.html", data)
}

func generatePostPages() []Post {
	posts, err := loadAllPosts()
	if err != nil {
		log.Fatal("Error loading posts:", err)
	}

	// Generate individual post pages
	for _, post := range posts {
		data := PageData{
			Title: post.Metadata.Title,
			Post:  &post,
		}

		outputPath := fmt.Sprintf("public/post/%s/index.html", post.Slug)
		os.MkdirAll(filepath.Dir(outputPath), 0755)
		renderToFile(outputPath, "post.html", data)
	}

	return posts
}

func generatePostsListPage(posts []Post) {
	data := PageData{
		Title: "Blog Posts",
		Posts: posts,
	}

	os.MkdirAll("public/posts", 0755)
	renderToFile("public/posts/index.html", "posts.html", data)
}

func renderToFile(outputPath, tmpl string, data interface{}) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, tmpl, data)
	if err != nil {
		log.Fatalf("Error rendering template %s: %v", tmpl, err)
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Error writing file %s: %v", outputPath, err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func copyBlogImages() {
	blogsDir := "blogs"
	entries, err := os.ReadDir(blogsDir)
	if err != nil {
		log.Printf("Error reading blogs directory: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		srcImages := filepath.Join(blogsDir, entry.Name(), "images")
		if _, err := os.Stat(srcImages); err == nil {
			dstImages := filepath.Join("public/post", entry.Name(), "images")
			os.MkdirAll(dstImages, 0755)
			copyDir(srcImages, dstImages)
		}
	}
}

// Reuse the same helper functions from main.go
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
