package handler

import (
	"encoding/json"
	"fmt"
	"imgpool/internal/config"
	"imgpool/internal/services/pool"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
)

// Image ...
type Image struct {
	ID     uint
	Width  uint
	Height uint
	Body   []byte
	Type   string
}

// Handler ...
type Handler struct {
	Router  *chi.Mux
	Service *pool.Service
	Config  *config.Config
	Process chan Image
}

// NewHandler ...
func NewHandler(config *config.Config, processChan chan Image, service *pool.Service) *Handler {
	return &Handler{
		Service: service,
		Config:  config,
		Process: processChan,
	}
}

// InitRoutes ...
func (h *Handler) InitRoutes() {
	h.Router = chi.NewRouter()
	h.Router.Post(h.Config.Server.Api+"image", h.Post)
	h.Router.Get(h.Config.Server.Api+"image/{id}", h.Get)
	h.Router.Get(h.Config.Server.Api+"image/download/{uri}", h.Download)
}

// Post ...
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Make sure the length is less than the capacity of the channel
	if len(h.Process) >= h.Config.Server.Queue {
		respondWithError(w, 500, "Try again next time", fmt.Sprintf("Service is buisy. Wait for %d seconds", h.Config.Server.Timeout))
		return
	}

	// We should take the data width & convert it to integer data type
	width, e := strconv.ParseUint(r.FormValue("width"), 10, 64)
	if e != nil {
		respondWithError(w, 400, "Unable to parse. Width should be an integer type", e.Error())
		return
	}

	// We should take the data height & convert it to integer data type
	height, e := strconv.ParseUint(r.FormValue("height"), 10, 64)
	if e != nil {
		respondWithError(w, 400, "Unable to parse. Height should be an integer type", e.Error())
		return
	}

	// Set rage of max size uploaded files
	e = r.ParseMultipartForm(10 << 20)
	if e != nil {
		respondWithError(w, 400, "The image size must not exceed 20 MB", e.Error())
		return
	}

	// We should take the data image & convert it to byte
	tmpImage, _, e := r.FormFile("image")
	if e != nil {
		respondWithError(w, 400, "The system cannot find the file specified", e.Error())
		return
	}

	defer tmpImage.Close() //Close after function return

	// The request body is an image that we want to do some fancy processing
	// on. That's hard work; we don't want to do too many of them at once, so
	// so we put those jobs in the small pool.

	b, e := ioutil.ReadAll(tmpImage)
	if e != nil {
		respondWithError(w, 500, "Unable to read data image", e.Error())
		return
	}

	contentType := http.DetectContentType(b)

	switch contentType {
	case "image/jpeg", "image/jpg":
	case "image/png":
		break
	default:
		respondWithError(w, 400, "Unsupported file type. Use .jpeg or .png instead of this", "Unsupported file type")
		return
	}

	propsImage := pool.Imgpool{
		Width:  uint(width),
		Height: uint(height),
	}

	// Push to database prepared properties & return new task.ID
	image, e := h.Service.Post(propsImage)
	if e != nil {
		respondWithError(w, 500, "Unable to post new image", e.Error())
		return
	}

	preparedImage := Image{
		ID:     image.ID,
		Width:  uint(width),
		Height: uint(height),
		Type:   contentType,
		Body:   b,
	}

	// Put 1 in the channel unless it is full
	h.Process <- preparedImage

	if e := json.NewEncoder(w).Encode(image); e != nil {
		panic(e)
	}
}

// Get ...
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := chi.URLParam(r, "id")

	taskId, e := strconv.ParseUint(id, 10, 64)
	if e != nil {
		respondWithError(w, 400, "Unable to parse id. Id should be an integer type", e.Error())
		return
	}

	task, e := h.Service.Get(uint(taskId))
	if e != nil {
		respondWithError(w, 404, "The image was not found or was deleted after expiration", e.Error())
		return
	}

	if e := json.NewEncoder(w).Encode(task); e != nil {
		panic(e)
	}
}

// Download ...
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	uri := chi.URLParam(r, "uri")
	if uri == "" {
		respondWithError(w, 404, "Get 'file' not specified in url.", "Get 'file' not specified in url.")
		return
	}

	//Check if file exists and open it
	openFile, e := os.Open(h.Config.Server.Path + uri)
	if e != nil {
		respondWithError(w, 404, "The image was not found or was deleted after expiration", e.Error())
		return
	}

	defer openFile.Close() //Close after function return

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	//Copy the headers into the FileHeader buffer
	openFile.Read(fileHeader)

	//Get content type of file
	fileContentType := http.DetectContentType(fileHeader)

	//Get the file size
	fileStat, _ := openFile.Stat()

	//Get file size as a string
	fileSize := strconv.FormatInt(fileStat.Size(), 10)

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+uri)
	w.Header().Set("Content-Type", fileContentType)
	w.Header().Set("Content-Length", fileSize)

	//We read 512 bytes from the file already, so we reset the offset back to 0
	openFile.Seek(0, 0)

	//'Copy' the file to the client
	io.Copy(w, openFile)
}

// Update ...
func (h *Handler) Update(id uint, tmpLink string) error {
	newLink := pool.Imgpool{
		Link: tmpLink,
	}
	if e := h.Service.Update(id, newLink); e != nil {
		return e
	}
	return nil
}

// Delete ...
func (h *Handler) Delete() {
	if e := h.Service.Delete(); e != nil {
		fmt.Printf("Unable to remove temporary file: %d\n", e)
		return
	}
}
