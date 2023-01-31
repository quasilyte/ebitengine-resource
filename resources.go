package resource

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"golang.org/x/image/font"
)

// AudioID is a typed key for Audio resources.
// See also: AudioInfo.
type AudioID int

type AudioInfo struct {
	// A path that will be used to read the resource data.
	Path string

	// Group is a sound group ID.
	// Groups are used to apply group-wide operations like
	// volume adjustments.
	// Conventionally, group 0 is "sound effect", 1 is "music", 2 is "voice".
	Group uint

	// Volume adjust how loud this sound will be.
	// The default value of 0 means "unadjusted".
	// Value greated than 0 increases the volume, negative values decrease it.
	// This setting accepts values in [-1, 1] range, where -1 mutes the sound
	// while 1 makes it as loud as possible.
	Volume float64
}

type Audio struct {
	// An ID that was associated with this resource.
	ID AudioID

	// An initialized audio player that can be used to play the audio.
	// Note that you may need to rewind it before playing the sound.
	// The player wraps an original stream, so you can't access it directly.
	Player *audio.Player

	Group  uint
	Volume float64
}

// FontID is a typed key for Font resources.
// See also: FontInfo.
type FontID int

type FontInfo struct {
	// A path that will be used to read the resource data.
	Path string

	Size int

	LineSpacing float64
}

type Font struct {
	// An ID that was associated with this resource.
	ID FontID

	Face font.Face
}

// ImageID is a typed key for Image resources.
// See also: ImageInfo.
type ImageID int

type ImageInfo struct {
	// A path that will be used to read the resource data.
	Path string

	FrameWidth  float64
	FrameHeight float64
}

type Image struct {
	// An ID that was associated with this resource.
	ID ImageID

	// An ebiten Image object initialized from the resource bytes.
	Data *ebiten.Image

	DefaultFrameWidth  float64
	DefaultFrameHeight float64
}

// RawID is a typed key for Raw resources.
// See also: RawInfo.
type RawID int

type RawInfo struct {
	// A path that will be used to read the resource data.
	Path string
}

type Raw struct {
	// An ID that was associated with this resource.
	ID RawID

	// Data is an uninterpreted resource contents.
	Data []byte
}

// ShaderID is a typed key for Shader resources.
// See also: ShaderInfo.
type ShaderID int

type ShaderInfo struct {
	// A path that will be used to read the resource data.
	Path string
}

type Shader struct {
	// An ID that was associated with this resource.
	ID ShaderID

	// A compiled shader.
	Data *ebiten.Shader
}
