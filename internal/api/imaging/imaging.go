package imaging

import (
	"image"

	"github.com/disintegration/imaging"
)

type Imaging interface {
	Open(path string) (image.Image, error)
	Blur(img image.Image, sigma float64) *image.NRGBA
	Save(img *image.NRGBA, path string) error
}

type ImagingImpl struct {
	Src image.Image
}

func NewImaging() Imaging {
	return &ImagingImpl{}
}

func (i *ImagingImpl) Open(path string) (image.Image, error) {
	src, err := imaging.Open(path)
	if err != nil {
		return nil, err
	}

	i.Src = src

	return i.Src, nil
}

func (i *ImagingImpl) Blur(img image.Image, sigma float64) *image.NRGBA {
	return imaging.Blur(img, sigma)
}

func (i *ImagingImpl) Save(img *image.NRGBA, path string) error {
	return imaging.Save(img, path)
}