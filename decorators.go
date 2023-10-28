package resource

import (
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// LoogOGG wraps OGG vorbis stream into an infinite loop.
func LoopOGG(stream io.ReadSeeker) io.ReadSeeker {
	oggStream := stream.(*vorbis.Stream)
	return audio.NewInfiniteLoop(oggStream, oggStream.Length())
}

// NopDecorator returns the input stream as is.
//
// This is only useful in combination with WAV resources
// that you don't wan't to eagerly load into memory.
// Using decorated WAVs can save some memory,
// but it will be more expensive to play them.
//
// See AudioInfo.StreamDecorator field comment to see the full story.
func NopDecorator(stream io.ReadSeeker) io.ReadSeeker {
	return stream
}
