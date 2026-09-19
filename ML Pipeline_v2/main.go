package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

const FastAPIServerURL = "http://localhost:8000/api/v1/verify-label"

func handleVerification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(64 << 20) // Limit memory max 64MB
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Transfer file streams to outbound multipart buffer
	files := r.MultipartForm.File["files"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		part, err := writer.CreateFormFile("files", fileHeader.Filename)
		if err != nil {
			file.Close()
			http.Error(w, "Error creating multipart file writer", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			http.Error(w, "Error copying binary content", http.StatusInternalServerError)
			return
		}
	}

	// Copy scalar inputs
	for key, values := range r.MultipartForm.Value {
		for _, val := range values {
			_ = writer.WriteField(key, val)
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", FastAPIServerURL, body)
	if err != nil {
		http.Error(w, "Failed to initialize request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "FastAPI Service connection error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func main() {
	// Serve index.html static assets
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/verify", handleVerification)

	log.Println("Go Server listening on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server launch failed: %s", err)
	}
}