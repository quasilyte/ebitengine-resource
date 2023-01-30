package resource

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Shader struct {
	ID ShaderID

	Data *ebiten.Shader
}

type ShaderInfo struct {
	Path string
}

type ShaderID int

type ShaderRegistry struct {
	mapping map[ShaderID]ShaderInfo
}

func (r *ShaderRegistry) Set(id ShaderID, info ShaderInfo) {
	r.mapping[id] = info
}

func (r *ShaderRegistry) Assign(m map[ShaderID]ShaderInfo) {
	for k, v := range m {
		r.Set(k, v)
	}
}
