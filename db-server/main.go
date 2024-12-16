package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type DBFile struct {
	Date int64  `json:"date"`
	URL  string `json:"url"`
}

type Server struct {
	dbDir    string
	dbFiles  []DBFile
	dbRegexp *regexp.Regexp
}

func NewServer(dbDir string) *Server {
	return &Server{
		dbDir:    dbDir,
		dbFiles:  make([]DBFile, 0),
		dbRegexp: regexp.MustCompile(`^db_(\d+)\.jsonl$`),
	}
}

func (s *Server) updateDBList() error {
	files, err := os.ReadDir(s.dbDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	newFiles := make([]DBFile, 0)

	for _, file := range files {
		if matches := s.dbRegexp.FindStringSubmatch(file.Name()); matches != nil {
			timestamp, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				log.Printf("Failed to parse timestamp from %s: %v", file.Name(), err)
				continue
			}

			fileURL := fmt.Sprintf("/db/%s", file.Name())
			newFiles = append(newFiles, DBFile{
				Date: timestamp,
				URL:  fileURL,
			})
		}
	}

	// Сортируем файлы по timestamp в убывающем порядке (новые первые)
	sort.Slice(newFiles, func(i, j int) bool {
		return newFiles[i].Date > newFiles[j].Date
	})

	s.dbFiles = newFiles
	return nil
}

func (s *Server) getLatestDB(w http.ResponseWriter, r *http.Request) {
	if err := s.updateDBList(); err != nil {
		http.Error(w, "Failed to update DB list", http.StatusInternalServerError)
		return
	}

	if len(s.dbFiles) == 0 {
		http.Error(w, "No DB files found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.dbFiles[0])
}

func (s *Server) serveDBFiles(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	if !s.dbRegexp.MatchString(filename) {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(s.dbDir, filename)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	http.ServeFile(w, r, filePath)
}

func main() {
	dbDir := flag.String("db-dir", ".", "Directory containing DB files")
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	server := NewServer(*dbDir)

	// Обновляем список файлов при запуске
	if err := server.updateDBList(); err != nil {
		log.Fatalf("Failed to initialize DB list: %v", err)
	}

	// API endpoint для получения последнего DB файла
	http.HandleFunc("/api/latest", server.getLatestDB)

	// Endpoint для отдачи файлов
	http.HandleFunc("/db/", server.serveDBFiles)

	log.Printf("Starting server on port %d...", *port)
	log.Printf("Serving DB files from: %s", *dbDir)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
