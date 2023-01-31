package resource

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type ShaderID int

type ShaderInfo struct {
	Path string
}

type Shader struct {
	ID ShaderID

	Data *ebiten.Shader
}
