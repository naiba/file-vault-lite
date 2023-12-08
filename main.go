package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
)

const (
	uploadDir   = "./uploads"
	maxFileSize = 1000 * 1024 * 1024 // 1000 MB
)

func humanReadableSize(i int64) string {
	var unit string
	var size float64

	switch {
	case i >= 1024*1024*1024:
		unit = "GB"
		size = float64(i) / (1024 * 1024 * 1024)
	case i >= 1024*1024:
		unit = "MB"
		size = float64(i) / (1024 * 1024)
	case i >= 1024:
		unit = "KB"
		size = float64(i) / 1024
	default:
		unit = "B"
		size = float64(i)
	}

	return fmt.Sprintf("%.2f %s", size, unit)
}

// ListFiles is a handler that begin with a "File Vault Lite" header and lists all files in the uploads directory, using ul and li tags with download link, show the file size in human readable and latest modified time.
func ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed")
		return
	}

	var files []os.FileInfo

	fmt.Fprintln(w, "<h1>File Vault Lite</h1>")
	fmt.Fprintln(w, "<ul>")
	err := filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if path == uploadDir {
			return nil
		}
		if err != nil {
			return err
		}
		files = append(files, info)
		return nil
	})

	slices.SortFunc(files, func(a, b fs.FileInfo) int {
		if a.ModTime().After(b.ModTime()) {
			return -1
		}
		if a.ModTime().Before(b.ModTime()) {
			return 1
		}
		return 0
	})

	for _, info := range files {
		fmt.Fprintf(w, "<li><a href=\"/download?filename=%s\">%s</a> (%s, %s)</li>", info.Name(), info.Name(), humanReadableSize(info.Size()), info.ModTime().Format("2006-01-02 15:04:05"))
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	fmt.Fprintln(w, "</ul>")
}

// BasicAuthMiddleware is a middleware that adds Basic Auth protection to a handler.
func BasicAuthMiddleware(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "Unauthorized")
			return
		}
		handler(w, r)
	}
}

// UploadHandler is a handler that handles file uploads.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed")
		return
	}

	r.ParseMultipartForm(int64(maxFileSize))
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer file.Close()

	filename := r.FormValue("filename")
	if filename == "" {
		filename = handler.Filename
	}
	filepath := filepath.Join(uploadDir, filename)

	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "File uploaded successfully.")
}

// DownloadHandler is a handler that handles file downloads.
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed")
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Filename not provided")
		return
	}
	filepath := filepath.Join(uploadDir, filename)

	f, err := os.Open(filepath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
}

func main() {
	// Define username and password for Basic Auth protection
	username := os.Getenv("FV_USERNAME")
	password := os.Getenv("FV_PASSWORD")

	log.Println("Starting server...")
	log.Printf("Username: %s\n", username)

	// Create an uploads directory if it does not exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	// Register handlers
	http.HandleFunc("/", ListFiles)
	http.HandleFunc("/upload", BasicAuthMiddleware(UploadHandler, username, password))
	http.HandleFunc("/download", BasicAuthMiddleware(DownloadHandler, username, password))

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
