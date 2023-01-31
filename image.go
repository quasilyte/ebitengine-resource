package resource

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type ImageID int

type ImageInfo struct {
	Path string

	FrameWidth  float64
	FrameHeight float64
}

type Image struct {
	ID   ImageID
	Data *ebiten.Image

	DefaultFrameWidth  float64
	DefaultFrameHeight float64
}
