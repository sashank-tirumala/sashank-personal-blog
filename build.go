//go:build build

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
	Snippet  template.HTML
	Intro    template.HTML
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
	// Clean output directory
	os.RemoveAll("public")
	os.MkdirAll("public", 0755)
	os.MkdirAll("public/post", 0755)
	os.MkdirAll("public/book", 0755)
	os.MkdirAll("public/books", 0755)

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
	books := generateBookPages()
	generateBooksListPage(books)

	// Copy blog images
	copyBlogImages()

	fmt.Println("Site built successfully in ./public")
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

func generateBookPages() []Book {
	books, err := loadAllBooks()
	if err != nil {
		log.Printf("Error loading books: %v", err)
		return nil
	}

	for _, book := range books {
		// Generate book table of contents page
		data := PageData{
			Title: book.Metadata.Title,
			Book:  &book,
		}

		bookDir := fmt.Sprintf("public/book/%s", book.Slug)
		os.MkdirAll(bookDir, 0755)
		renderToFile(filepath.Join(bookDir, "index.html"), "book.html", data)

		// Copy EPUB file if it exists
		if book.Metadata.EpubFile != "" {
			srcEpub := filepath.Join("books", book.Slug, book.Metadata.EpubFile)
			dstEpub := filepath.Join(bookDir, book.Metadata.EpubFile)
			if err := copyFile(srcEpub, dstEpub); err != nil {
				log.Printf("Error copying EPUB for %s: %v", book.Slug, err)
			} else {
				fmt.Printf("Copied: %s\n", dstEpub)
			}
		}

		// Generate individual chapter pages
		for i, chapterInfo := range book.Chapters {
			chapter, err := loadChapter(&book, chapterInfo.Slug)
			if err != nil {
				log.Printf("Error loading chapter %s: %v", chapterInfo.Slug, err)
				continue
			}

			// Set prev/next chapters
			if i > 0 {
				chapter.PrevChapter = &book.Chapters[i-1]
			}
			if i < len(book.Chapters)-1 {
				chapter.NextChapter = &book.Chapters[i+1]
			}

			chapterData := PageData{
				Title:   chapter.Title + " - " + book.Metadata.Title,
				Book:    &book,
				Chapter: chapter,
			}

			chapterPath := filepath.Join(bookDir, chapterInfo.Slug, "index.html")
			os.MkdirAll(filepath.Dir(chapterPath), 0755)
			renderToFile(chapterPath, "chapter.html", chapterData)
		}
	}

	return books
}

func generateBooksListPage(books []Book) {
	data := PageData{
		Title: "Books",
		Books: books,
	}

	renderToFile("public/books/index.html", "books.html", data)
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

	// Read snippet (optional) - short intro for books list
	var snippet template.HTML
	snippetPath := filepath.Join(bookDir, "snippet.html")
	if snippetData, err := os.ReadFile(snippetPath); err == nil {
		snippet = template.HTML(snippetData)
	}

	// Read intro (optional) - longer intro for book page
	var intro template.HTML
	introPath := filepath.Join(bookDir, "intro.html")
	if introData, err := os.ReadFile(introPath); err == nil {
		intro = template.HTML(introData)
	}

	return &Book{
		Metadata: metadata,
		Slug:     slug,
		Chapters: chaptersConfig.Chapters,
		Snippet:  snippet,
		Intro:    intro,
	}, nil
}

func loadChapter(book *Book, chapterSlug string) (*ChapterData, error) {
	// Find chapter info
	var chapterInfo ChapterInfo
	for _, ch := range book.Chapters {
		if ch.Slug == chapterSlug {
			chapterInfo = ch
			break
		}
	}

	// Read chapter content
	chapterPath := filepath.Join("books", book.Slug, "chapters", chapterSlug+".xhtml")
	content, err := os.ReadFile(chapterPath)
	if err != nil {
		return nil, err
	}

	return &ChapterData{
		Title:       chapterInfo.Title,
		Content:     template.HTML(content),
		BookSlug:    book.Slug,
		BookTitle:   book.Metadata.Title,
		ChapterSlug: chapterSlug,
	}, nil
}
