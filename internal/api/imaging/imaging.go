package imaging

import (
	"image"

	"github.com/disintegration/imaging"
)

type Imaging struct {
	Src image.Image
}

func NewImaging() *Imaging {
	return &Imaging{}
}

func (i *Imaging) Open(path string) (image.Image, error) {
	src, err := imaging.Open(path)
	if err != nil {
		return nil, err
	}

	i.Src = src

	return i.Src, nil
}

func (i *Imaging) Blur(img image.Image, sigma float64) *image.NRGBA {
	return imaging.Blur(img, sigma)
}

func (i *Imaging) Save(img *image.NRGBA, path string) error {
	return imaging.Save(img, path)
}