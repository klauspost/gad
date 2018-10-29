package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	dir, _ := os.Getwd()
	if !filepath.IsAbs(dir) {
		dir, _ = filepath.Abs(dir)
	}
	dir = filepath.Clean(dir)
	fmt.Println("Serving from:", dir, "on http://localhost:8080")
	http.Handle("/", http.FileServer(http.Dir(dir)))
	log.Print(http.ListenAndServe(":8080", nil))
}
