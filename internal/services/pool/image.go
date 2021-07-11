package pool

import (
	"fmt"
	"imgpool/internal/config"
	"os"
	"strconv"

	"gorm.io/gorm"
)

// Service ...
type Service struct {
	Conn   *gorm.DB
	Config *config.Config
}

// Imgpool ...
type Imgpool struct {
	gorm.Model
	Link   string
	Width  uint
	Height uint
}

// ImgpoolService ...
type ImgpoolService interface {
	PostImage(image Imgpool) (Imgpool, error)
	GetImage(ID uint) (Imgpool, error)
	UpdateImage(ID uint, newLink Imgpool) error
	DeleteImage(ID uint) error
}

// NewService ...
func NewService(conn *gorm.DB, config *config.Config) *Service {
	return &Service{
		Conn:   conn,
		Config: config,
	}
}

// Post ...
func (s *Service) Post(image Imgpool) (Imgpool, error) {
	if r := s.Conn.Save(&image); r.Error != nil {
		return Imgpool{}, r.Error
	}

	return image, nil
}

// Get ...
func (s *Service) Get(ID uint) (Imgpool, error) {
	var image Imgpool

	r := s.Conn.First(&image, ID)
	if r.Error != nil {
		return Imgpool{}, r.Error
	}

	return image, nil
}

// Update ...
func (s *Service) Update(ID uint, newLink Imgpool) error {
	image, e := s.Get(ID)
	if e != nil {
		return e
	}

	if r := s.Conn.Model(&image).Updates(newLink); r.Error != nil {
		return r.Error
	}

	return nil
}

// Delete ...
func (s *Service) Delete() error {
	var (
		images  []Imgpool
		timeout = strconv.Itoa(s.Config.Server.Timeout)
	)

	if r := s.Conn.Raw(fmt.Sprintf("Select * from imgpools where updated_at + interval '%s seconds' < current_timestamp", timeout)).Scan(&images); r.Error != nil {
		return r.Error
	}

	go func() {
		for _, i := range images {
			if e := os.Remove(s.Config.Server.Path + i.Link); e != nil {
				fmt.Printf("Unable to remove temporary file: %d\n", e)
			}
			if r := s.Conn.Unscoped().Delete(&Imgpool{}, i.ID); r.Error != nil {
				fmt.Printf("Unable to remove row: %d\n", r.Error)
				return
			}

			fmt.Printf("File %s was successfully removed\n", i.Link)
		}
	}()

	return nil
}
