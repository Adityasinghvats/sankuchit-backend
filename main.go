package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
)

func resizeCompressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20) //limit to 10 mb
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}
	//get file from that we will find image
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	targetWidthStr := r.FormValue("width")
	targetHeightStr := r.FormValue("height")
	qualityStr := r.FormValue("quality")
	format := r.FormValue("format")
	//set default format
	if format == "" {
		format = "jpeg"
	}
	//get height , width and quality
	targetWidth, err := strconv.ParseUint(targetWidthStr, 10, 32) //base 10 decimal which is a 32 bit integer
	if err != nil {
		http.Error(w, "Invalid width value", http.StatusBadRequest)
		return
	}
	targetHeight, err := strconv.ParseUint(targetHeightStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid height value", http.StatusBadRequest)
		return
	}
	quality, err := strconv.ParseUint(qualityStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid quality value", http.StatusBadRequest)
		return
	}
	image, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}
	resizedImage := resize.Resize(uint(targetWidth), uint(targetHeight), image, resize.Lanczos3)
	var outputBytes bytes.Buffer
	switch strings.ToLower((format)) {
	case "jpeg":
		err = jpeg.Encode(&outputBytes, resizedImage, &jpeg.Options{Quality: int(quality)})
		if err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=resized_image.jpg")
	case "png":
		err = png.Encode(&outputBytes, resizedImage)
		if err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename=resized_image.png")
	default:
		http.Error(w, "Invalid format", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(outputBytes.Len()))
	w.Write(outputBytes.Bytes())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/resize", resizeCompressHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Starting server at port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
