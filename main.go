package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// PageData represents data passed to templates
type PageData struct {
	Title   string
	Content string
}

func main() {
	// Parse templates
	tmpl, err := template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// Define routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "My Personal Website",
			Content: "Welcome to my minimalist blog!",
		}

		err := tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
		}
	})

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
