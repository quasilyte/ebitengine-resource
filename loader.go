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

	// CustomAudioLoader allows LoadAudio to load audio formats that are not supported by default.
	// If it's nil, LoadAudio() will support only ".ogg" and ".wav" formats.
	//
	// CustomAudioLoader should load the audio resource in a form that is suitable for
	// the Ebitengine audio.NewPlayer() argument.
	// Note: the input reader will be closed as soon as resource loading is finished.
	// If your resource needs it to stay valid, create its copy.
	//
	// This function should return nil if it can't handle a given resource.
	//
	// It's called exactly once per every unique AudioID being loaded.
	// Setting this field back to nil after all custom audio resources are loaded
	// will still keep loaded resources reachable via LoadAudio(AudioID).
	// If your game uses a simple preload-everything scheme, you might want to
	// set this field to nil after you're done with preloading.
	//
	// Keep in mind that you have to use LoadAudio instead of LoadOGG or LoadWAV to
	// fetch the custom resources.
	//
	// An example of this function is XM (or MP3) loading routine.
	// It would check the filename for ".xm" suffix, read the data from r and
	// produce an XM stream out of it.
	//
	// You can't use this function to override the way OGG or WAV is being loaded
	// as this function is called after the default loaders and it's by design.
	CustomAudioLoader func(r io.Reader, info AudioInfo) io.ReadSeeker

	ImageRegistry  registry[ImageID, ImageInfo]
	AudioRegistry  registry[AudioID, AudioInfo]
	FontRegistry   registry[FontID, FontInfo]
	ShaderRegistry registry[ShaderID, ShaderInfo]
	RawRegistry    registry[RawID, RawInfo]

	audioContext *audio.Context

	images      map[ImageID]Image
	shaders     map[ShaderID]Shader
	wavs        map[AudioID]Audio
	oggs        map[AudioID]Audio
	customAudio map[AudioID]Audio
	fonts       map[FontID]Font
	raws        map[RawID]Raw
}

// NewLoader creates a new resources loader that serves as both
// resource accessor and decoded resources cache.
//
// An audio context is required to enable audio-related code to work.
// Audio resources are cached as *audio.Players and they can't
// be created without an initialized Ebitengine audio context.
func NewLoader(audioContext *audio.Context) *Loader {
	l := &Loader{
		images:      make(map[ImageID]Image),
		shaders:     make(map[ShaderID]Shader),
		wavs:        make(map[AudioID]Audio),
		oggs:        make(map[AudioID]Audio),
		customAudio: make(map[AudioID]Audio),
		fonts:       make(map[FontID]Font),
		raws:        make(map[RawID]Raw),
	}
	l.audioContext = audioContext
	l.AudioRegistry.mapping = make(map[AudioID]AudioInfo)
	l.ImageRegistry.mapping = make(map[ImageID]ImageInfo)
	l.ShaderRegistry.mapping = make(map[ShaderID]ShaderInfo)
	l.FontRegistry.mapping = make(map[FontID]FontInfo)
	l.RawRegistry.mapping = make(map[RawID]RawInfo)
	return l
}

// LoadAudio is a helper method that will use an appropriate
// Load method depending on the filename extension.
//
// For example, it will use LoadOGG for ".ogg" files.
func (l *Loader) LoadAudio(id AudioID) Audio {
	audioInfo := l.getAudioInfo(id)
	if strings.HasSuffix(audioInfo.Path, ".ogg") {
		return l.LoadOGG(id)
	}
	if strings.HasSuffix(audioInfo.Path, ".wav") {
		return l.LoadWAV(id)
	}
	if len(l.customAudio) != 0 || l.CustomAudioLoader != nil {
		// Even if CustomAudioLoader is nil at this point, we might still have
		// cached custom audio resources.
		// Let them a chance to be fetched.
		a, ok := l.loadCustomAudio(id, audioInfo)
		if ok {
			return a
		}
	}
	panic(fmt.Sprintf("load %q audio: unrecognized format", audioInfo.Path))
}

// GetFontInfo extracts the audio info associated with a given key.
func (l *Loader) GetAudioInfo(id AudioID) AudioInfo {
	return l.AudioRegistry.mapping[id]
}

// LoadWAV returns an Audio resource associated with a given key.
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
		stream, err := wav.DecodeWithoutResampling(r)
		if err != nil {
			panic(fmt.Sprintf("decode %q wav: %v", wavInfo.Path, err))
		}
		var player *audio.Player
		if wavInfo.StreamDecorator == nil {
			// Good, can read it into the memory.
			wavData := make([]byte, stream.Length())
			if _, err := io.ReadFull(stream, wavData); err != nil {
				panic(fmt.Sprintf("read %q wav: %v", wavInfo.Path, err))
			}
			player = l.audioContext.NewPlayerFromBytes(wavData)
		} else {
			// This is an explicit way to tell "don't read it into the memory".
			// Also, some streams can have external dependencies to affect the
			// sound, so we can't rely on the bytes being the same every time.
			player, err = l.audioContext.NewPlayer(wavInfo.StreamDecorator(stream))
			if err != nil {
				panic(err.Error())
			}
		}
		a = l.createAudioObject(player, id, wavInfo)
		l.wavs[id] = a
	}
	return a
}

// LoadOGG returns an Audio resource associated with a given key.
// Only a first call for this id will lead to resource decoding,
// all next calls return the cached result.
func (l *Loader) LoadOGG(id AudioID) Audio {
	a, ok := l.oggs[id]
	if !ok {
		oggInfo := l.getAudioInfo(id)
		// Do not close this reader as it would break the stream with "file already closed".
		r := l.OpenAssetFunc(oggInfo.Path)
		var err error
		oggStream, err := vorbis.DecodeWithoutResampling(r)
		if err != nil {
			panic(fmt.Sprintf("decode %q ogg: %v", oggInfo.Path, err))
		}
		player, err := l.audioContext.NewPlayer(l.maybeWrapAudioStream(oggStream, oggInfo))
		if err != nil {
			panic(err.Error())
		}
		a = l.createAudioObject(player, id, oggInfo)
		l.oggs[id] = a
	}
	return a
}

func (l *Loader) loadCustomAudio(id AudioID, info AudioInfo) (Audio, bool) {
	a, ok := l.customAudio[id]
	if !ok {
		if l.CustomAudioLoader == nil {
			// Can't load a new custom audio resource without this function.
			return a, false
		}
		r := l.OpenAssetFunc(info.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q custom audio reader: %v", info.Path, err))
			}
		}()
		stream := l.CustomAudioLoader(r, info)
		if stream == nil {
			return a, false
		}
		player, err := l.audioContext.NewPlayer(l.maybeWrapAudioStream(stream, info))
		if err != nil {
			panic(err.Error())
		}
		a = l.createAudioObject(player, id, info)
		l.customAudio[id] = a
	}
	return a, true
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

// GetFontInfo extracts the font info associated with a given key.
func (l *Loader) GetFontInfo(id FontID) FontInfo {
	return l.FontRegistry.mapping[id]
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

// GetImageInfo extracts the image info associated with a given key.
func (l *Loader) GetImageInfo(id ImageID) ImageInfo {
	return l.ImageRegistry.mapping[id]
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

// GetRawInfo extracts the raw info associated with a given key.
func (l *Loader) GetRawInfo(id RawID) RawInfo {
	return l.RawRegistry.mapping[id]
}

func (l *Loader) getAudioInfo(id AudioID) AudioInfo {
	info, ok := l.AudioRegistry.mapping[id]
	if !ok {
		panic(fmt.Sprintf("unregistered audio with id=%d", id))
	}
	return info
}

func (l *Loader) createAudioObject(p *audio.Player, id AudioID, info AudioInfo) Audio {
	volume := (info.Volume / 2) + 0.5
	return Audio{
		ID:     id,
		Player: p,
		Volume: volume,
		Group:  info.Group,
	}
}

func (l *Loader) maybeWrapAudioStream(r io.ReadSeeker, info AudioInfo) io.ReadSeeker {
	if info.StreamDecorator != nil {
		return info.StreamDecorator(r)
	}
	return r
}
