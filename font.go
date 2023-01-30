package resource

import (
	"golang.org/x/image/font"
)

type Font struct {
	ID FontID

	Face font.Face
}

type FontInfo struct {
	Path string

	Size int

	LineSpacing float64
}

type FontID int

type FontRegistry struct {
	mapping map[FontID]FontInfo
}

func (r *FontRegistry) Set(id FontID, info FontInfo) {
	r.mapping[id] = info
}

func (r *FontRegistry) Assign(m map[FontID]FontInfo) {
	for k, v := range m {
		r.Set(k, v)
	}
}
