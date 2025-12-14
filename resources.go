package resource

import (
	"io"

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

	// StreamDecorator is a way to wrap resource stream into another stream
	// before the associated audio player is created.
	//
	// An example usage for this is to wrap OGG stream into an InfiniteLoop stream.
	// You can use LoopOGG function just for that.
	// Another example could include a preprocessing stream that would
	// alter the original resource sound.
	//
	// Previously, OGG streams were looped by default.
	// This option allows a finer control over what to do with a stream.
	//
	// A nil decorator will use the original stream as is.
	//
	// This function is called exactly once per resource upon its first loading.
	//
	// Note: you can use decorators for any type of audio resource (WAV included),
	// although it's not recommended to do so. By default, this library reads the
	// entire WAV into memory and creates the raw byte stream player from that.
	// Unless you really want to do it, leave this field nil for WAVs.
	// If you want to reduce the application memory footprint, it can be
	// beneficial to add a NopDecorator decorator that would return the input stream as is.
	// This will make WAV more expensive to play in terms of CPU clocks.
	StreamDecorator func(stream io.ReadSeeker) io.ReadSeeker
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

	// For some formats (e.g. wav) this value will hold a duration in secods.
	// If it's 0, then this value can not be trusted.
	Duration float64
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

	FrameWidth  int
	FrameHeight int
}

type Image struct {
	// An ID that was associated with this resource.
	ID ImageID

	// An ebiten Image object initialized from the resource bytes.
	Data *ebiten.Image

	DefaultFrameWidth  int
	DefaultFrameHeight int
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
