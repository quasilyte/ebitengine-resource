package resource

import (
	"fmt"
	"image"
	"io"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// Loader is used to load and cache game resources like images and audio files.
type Loader struct {
	// OpenAssetFunc is used to open an asset resource identified by its path.
	// The returned resource will be closed after it will be loaded.
	OpenAssetFunc func(path string) io.ReadCloser

	ImageRegistry  registry[ImageID, ImageInfo]
	AudioRegistry  registry[AudioID, AudioInfo]
	FontRegistry   registry[FontID, FontInfo]
	ShaderRegistry registry[ShaderID, ShaderInfo]
	RawRegistry    registry[RawID, RawInfo]

	audioContext *audio.Context

	images  map[ImageID]Image
	shaders map[ShaderID]Shader
	wavs    map[AudioID]Audio
	oggs    map[AudioID]Audio
	fonts   map[FontID]Font
	raws    map[RawID]Raw
}

// NewLoader creates a new resources loader that serves as both
// resource accessor and decoded resources cache.
//
// An audio context is required to enable audio-related code to work.
// Audio resources are cached as *audio.Players and they can't
// be created without an initialized Ebitengine audio context.
func NewLoader(audioContext *audio.Context) *Loader {
	l := &Loader{
		images:  make(map[ImageID]Image),
		shaders: make(map[ShaderID]Shader),
		wavs:    make(map[AudioID]Audio),
		oggs:    make(map[AudioID]Audio),
		fonts:   make(map[FontID]Font),
		raws:    make(map[RawID]Raw),
	}
	l.audioContext = audioContext
	l.AudioRegistry.mapping = make(map[AudioID]AudioInfo)
	l.ImageRegistry.mapping = make(map[ImageID]ImageInfo)
	l.ShaderRegistry.mapping = make(map[ShaderID]ShaderInfo)
	l.FontRegistry.mapping = make(map[FontID]FontInfo)
	l.RawRegistry.mapping = make(map[RawID]RawInfo)
	return l
}

// LoadAudio is a helper method that will use an appripriate
// Load method depending on the filename extension.
// For example, it will use LoadOGG for ".ogg" files.
func (l *Loader) LoadAudio(id AudioID) Audio {
	audioInfo := l.getAudioInfo(id)
	if strings.HasSuffix(audioInfo.Path, ".ogg") {
		return l.LoadOGG(id)
	}
	return l.LoadWAV(id)
}

// LoadWAV returns a Audio resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadWAV(id AudioID) Audio {
	a, ok := l.wavs[id]
	if !ok {
		wavInfo := l.getAudioInfo(id)
		r := l.OpenAssetFunc(wavInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q wav reader: %v", wavInfo.Path, err))
			}
		}()
		stream, err := wav.DecodeWithSampleRate(l.audioContext.SampleRate(), r)
		if err != nil {
			panic(fmt.Sprintf("decode %q wav: %v", wavInfo.Path, err))
		}
		wavData, err := io.ReadAll(stream)
		if err != nil {
			panic(fmt.Sprintf("read %q wav: %v", wavInfo.Path, err))
		}
		player := l.audioContext.NewPlayerFromBytes(wavData)
		volume := (wavInfo.Volume / 2) + 0.5
		a = Audio{
			ID:     id,
			Player: player,
			Volume: volume,
			Group:  wavInfo.Group,
		}
		l.wavs[id] = a
	}
	return a
}

// LoadOGG returns a Audio resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadOGG(id AudioID) Audio {
	a, ok := l.oggs[id]
	if !ok {
		oggInfo := l.getAudioInfo(id)
		r := l.OpenAssetFunc(oggInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q ogg reader: %v", oggInfo.Path, err))
			}
		}()
		var err error
		oggStream, err := vorbis.DecodeWithSampleRate(l.audioContext.SampleRate(), r)
		if err != nil {
			panic(fmt.Sprintf("decode %q ogg: %v", oggInfo.Path, err))
		}
		loopedStream := audio.NewInfiniteLoop(oggStream, oggStream.Length())
		player, err := l.audioContext.NewPlayer(loopedStream)
		if err != nil {
			panic(err.Error())
		}
		volume := (oggInfo.Volume / 2) + 0.5
		a = Audio{
			ID:     id,
			Player: player,
			Volume: volume,
			Group:  oggInfo.Group,
		}
		l.oggs[id] = a
	}
	return a
}

// LoadFont returns a Font resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadFont(id FontID) Font {
	f, ok := l.fonts[id]
	if !ok {
		fontInfo, ok := l.FontRegistry.mapping[id]
		if !ok {
			panic(fmt.Sprintf("unregistered font with id=%d", id))
		}
		r := l.OpenAssetFunc(fontInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q font reader: %v", fontInfo.Path, err))
			}
		}()
		fontData, err := io.ReadAll(r)
		if err != nil {
			panic(fmt.Sprintf("reading %q data: %v", fontInfo.Path, err))
		}
		tt, err := opentype.Parse(fontData)
		if err != nil {
			panic(fmt.Sprintf("parsing %q font: %v", fontInfo.Path, err))
		}
		face, err := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    float64(fontInfo.Size),
			DPI:     96,
			Hinting: font.HintingFull,
		})
		if err != nil {
			panic(fmt.Sprintf("creating a font face for %q: %v", fontInfo.Path, err))
		}
		if fontInfo.LineSpacing != 0 && fontInfo.LineSpacing != 1 {
			h := float64(face.Metrics().Height.Round()) * fontInfo.LineSpacing
			face = text.FaceWithLineHeight(face, math.Round(h))
		}
		f = Font{
			ID:   id,
			Face: face,
		}
		l.fonts[id] = f
	}
	return f
}

// LoadImage returns an Image resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadImage(id ImageID) Image {
	img, ok := l.images[id]
	if !ok {
		imageInfo, ok := l.ImageRegistry.mapping[id]
		if !ok {
			panic(fmt.Sprintf("unregistered image with id=%d", id))
		}
		r := l.OpenAssetFunc(imageInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q image reader: %v", imageInfo.Path, err))
			}
		}()
		rawImage, _, err := image.Decode(r)
		if err != nil {
			panic(fmt.Sprintf("decode %q image: %v", imageInfo.Path, err))
		}
		data := ebiten.NewImageFromImage(rawImage)
		img = Image{
			ID:                 id,
			Data:               data,
			DefaultFrameWidth:  imageInfo.FrameWidth,
			DefaultFrameHeight: imageInfo.FrameHeight,
		}
		l.images[id] = img
	}
	return img
}

// LoadShader returns a Shader resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadShader(id ShaderID) Shader {
	shader, ok := l.shaders[id]
	if !ok {
		shaderInfo, ok := l.ShaderRegistry.mapping[id]
		if !ok {
			panic(fmt.Sprintf("unregistered shader with id=%d", id))
		}
		r := l.OpenAssetFunc(shaderInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q shader reader: %v", shaderInfo.Path, err))
			}
		}()
		data, err := io.ReadAll(r)
		if err != nil {
			panic(fmt.Sprintf("read %q shader: %v", shaderInfo.Path, err))
		}
		rawShader, err := ebiten.NewShader(data)
		if err != nil {
			panic(fmt.Sprintf("compile %q shader: %v", shaderInfo.Path, err))
		}
		shader = Shader{
			ID:   id,
			Data: rawShader,
		}
		l.shaders[id] = shader
	}
	return shader
}

// LoadRaw returns a Raw resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadRaw(id RawID) Raw {
	raw, ok := l.raws[id]
	if !ok {
		rawInfo, ok := l.RawRegistry.mapping[id]
		if !ok {
			panic(fmt.Sprintf("unregistered raw with id=%d", id))
		}
		r := l.OpenAssetFunc(rawInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q raw reader: %v", rawInfo.Path, err))
			}
		}()
		data, err := io.ReadAll(r)
		if err != nil {
			panic(fmt.Sprintf("read %q raw: %v", rawInfo.Path, err))
		}
		raw = Raw{
			ID:   id,
			Data: data,
		}
		l.raws[id] = raw
	}
	return raw
}

func (l *Loader) getAudioInfo(id AudioID) AudioInfo {
	info, ok := l.AudioRegistry.mapping[id]
	if !ok {
		panic(fmt.Sprintf("unregistered audio with id=%d", id))
	}
	return info
}
