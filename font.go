package resource

import (
	"golang.org/x/image/font"
)

type FontID int

type FontInfo struct {
	Path string

	Size int

	LineSpacing float64
}

type Font struct {
	ID FontID

	Face font.Face
}
