package process

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"imgpool/internal/config"
	"imgpool/internal/handler"
	"imgpool/internal/services/pool"
	"math/rand"
	"os"
	"time"

	"github.com/nfnt/resize"
)

// ProcessImage
func ProcessImage(c *config.Config, task *handler.Image) (pool.Imgpool, error) {
	// 	Hard work emulate
	time.Sleep(time.Duration(c.Server.Duration) * time.Second)

	// Decode from byte
	decodedBody, _, e := image.Decode(bytes.NewReader(task.Body))
	if e != nil {
		fmt.Printf("Unable to decode image: %s\n", e)
		return pool.Imgpool{}, e
	}

	// Resize decodedBody
	tmpImage := resize.Resize(task.Width, task.Height, decodedBody, resize.Lanczos3)

	// Creating a temporary file name
	char := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	hash := make([]rune, c.Server.Hash)

	for i := range hash {
		hash[i] = char[rand.Intn(len(char))]
	}

	extImage := ContentType(task.Type)

	tmpLink := string(hash) + fmt.Sprint(task.ID) + extImage

	// Move file to temporary directory
	output, e := os.Create(c.Server.Path + tmpLink)
	if e != nil {
		fmt.Printf("Output dirrectory not found: %s\n", e)
		return pool.Imgpool{}, e
	}

	defer output.Close() //Close after function return

	// Encode image to .jpeg type
	png.Encode(output, tmpImage)

	return pool.Imgpool{Link: tmpLink}, nil
}

// ContentType
func ContentType(t string) string {
	switch t {
	case "image/jpeg", "image/jpg":
		return ".jpeg"
	case "image/png":
		return ".png"
	default:
		fmt.Printf("The type of image is unknown\n")
	}
	return t
}
