//go:build !build

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

// BookMetadata represents the metadata for a book
type BookMetadata struct {
	Title       string `yaml:"title"`
	Subtitle    string `yaml:"subtitle"`
	Author      string `yaml:"author"`
	Translator  string `yaml:"translator"`
	Editor      string `yaml:"editor"`
	Illustrator string `yaml:"illustrator"`
	Year        int    `yaml:"year"`
	Description string `yaml:"description"`
	EpubFile    string `yaml:"epub_file"`
}

// ChapterInfo represents a chapter entry in chapters.yaml
type ChapterInfo struct {
	Slug  string `yaml:"slug"`
	Title string `yaml:"title"`
}

// ChaptersConfig represents the chapters.yaml structure
type ChaptersConfig struct {
	Chapters []ChapterInfo `yaml:"chapters"`
}

// Book represents a complete book
type Book struct {
	Metadata BookMetadata
	Slug     string
	Chapters []ChapterInfo
}

// ChapterData represents data for rendering a chapter
type ChapterData struct {
	Title       string
	Content     template.HTML
	BookSlug    string
	BookTitle   string
	ChapterSlug string
	PrevChapter *ChapterInfo
	NextChapter *ChapterInfo
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
	Books   []Book
	Book    *Book
	Chapter *ChapterData
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
	http.HandleFunc("/books/", booksHandler)
	http.HandleFunc("/book/", bookHandler)

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

func booksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := loadAllBooks()
	if err != nil {
		http.Error(w, "Error loading books", http.StatusInternalServerError)
		log.Printf("Error loading books: %v", err)
		return
	}

	data := PageData{
		Title: "Books",
		Books: books,
	}

	renderTemplate(w, "books.html", data)
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	// URL format: /book/{bookSlug}/ or /book/{bookSlug}/{chapterSlug} or /book/{bookSlug}/download
	path := strings.TrimPrefix(r.URL.Path, "/book/")
	path = strings.TrimSuffix(path, "/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	bookSlug := parts[0]

	// Load the book
	book, err := loadBook(bookSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// If only book slug, show table of contents
	if len(parts) == 1 || parts[1] == "" {
		data := PageData{
			Title: book.Metadata.Title,
			Book:  book,
		}
		renderTemplate(w, "book.html", data)
		return
	}

	// Handle EPUB download
	if strings.HasSuffix(parts[1], ".epub") && parts[1] == book.Metadata.EpubFile {
		epubPath := filepath.Join("books", bookSlug, book.Metadata.EpubFile)
		w.Header().Set("Content-Type", "application/epub+zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+book.Metadata.EpubFile+"\"")
		http.ServeFile(w, r, epubPath)
		return
	}

	// Otherwise, show the chapter
	chapterSlug := parts[1]
	chapter, err := loadChapter(book, chapterSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := PageData{
		Title:   chapter.Title + " - " + book.Metadata.Title,
		Book:    book,
		Chapter: chapter,
	}

	renderTemplate(w, "chapter.html", data)
}

func loadAllBooks() ([]Book, error) {
	var books []Book

	booksDir := "books"
	entries, err := os.ReadDir(booksDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		book, err := loadBook(entry.Name())
		if err != nil {
			log.Printf("Error loading book %s: %v", entry.Name(), err)
			continue
		}

		books = append(books, *book)
	}

	return books, nil
}

func loadBook(slug string) (*Book, error) {
	bookDir := filepath.Join("books", slug)

	// Read metadata
	metadataPath := filepath.Join(bookDir, "metadata.yaml")
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata BookMetadata
	err = yaml.Unmarshal(metadataData, &metadata)
	if err != nil {
		return nil, err
	}

	// Read chapters config
	chaptersPath := filepath.Join(bookDir, "chapters.yaml")
	chaptersData, err := os.ReadFile(chaptersPath)
	if err != nil {
		return nil, err
	}

	var chaptersConfig ChaptersConfig
	err = yaml.Unmarshal(chaptersData, &chaptersConfig)
	if err != nil {
		return nil, err
	}

	return &Book{
		Metadata: metadata,
		Slug:     slug,
		Chapters: chaptersConfig.Chapters,
	}, nil
}

func loadChapter(book *Book, chapterSlug string) (*ChapterData, error) {
	// Find chapter index
	chapterIndex := -1
	var chapterInfo ChapterInfo
	for i, ch := range book.Chapters {
		if ch.Slug == chapterSlug {
			chapterIndex = i
			chapterInfo = ch
			break
		}
	}

	if chapterIndex == -1 {
		return nil, os.ErrNotExist
	}

	// Read chapter content
	chapterPath := filepath.Join("books", book.Slug, "chapters", chapterSlug+".xhtml")
	content, err := os.ReadFile(chapterPath)
	if err != nil {
		return nil, err
	}

	// Determine prev/next chapters
	var prevChapter, nextChapter *ChapterInfo
	if chapterIndex > 0 {
		prevChapter = &book.Chapters[chapterIndex-1]
	}
	if chapterIndex < len(book.Chapters)-1 {
		nextChapter = &book.Chapters[chapterIndex+1]
	}

	return &ChapterData{
		Title:       chapterInfo.Title,
		Content:     template.HTML(content),
		BookSlug:    book.Slug,
		BookTitle:   book.Metadata.Title,
		ChapterSlug: chapterSlug,
		PrevChapter: prevChapter,
		NextChapter: nextChapter,
	}, nil
}
