package resource_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	resource "github.com/quasilyte/ebitengine-resource"
)

// The essential game data resources could be enumerated using iota constants.
// Dynamic game content requires dynamically generated IDs during the run time.
const (
	rawNone resource.RawID = iota
	rawLevel1Data
	rawLevel2Data
	rawDefaultConfig
)

const (
	audioNone resource.AudioID = iota
	audioExample
)

func Example() {
	audioContext := audio.NewContext(44100)

	l := resource.NewLoader(audioContext)

	l.OpenAssetFunc = func(path string) io.ReadCloser {
		return io.NopCloser(bytes.NewReader(resdata[path]))
	}

	// Before any resource is loadable, they should be bound.
	// You bind (register) resources by using typed registries.
	// For instance, RawRegistry is used to register Raw resources.
	rawResources := map[resource.RawID]resource.RawInfo{
		rawLevel1Data:    {Path: "maps/level1.json"},
		rawLevel2Data:    {Path: "maps/level2.json"},
		rawDefaultConfig: {Path: "config.txt"},
	}
	l.RawRegistry.Assign(rawResources)
	l.AudioRegistry.Assign(map[resource.AudioID]resource.AudioInfo{
		audioExample: {Path: "audio/example.wav", Volume: -0.2},
	})

	// It's possible to preload the resources.
	// Just load them once during the load screen or game initialization.
	// The second Load for the same resource would return a cached result.
	for id := range rawResources {
		l.LoadRaw(id)
	}

	// Raw resources are stored as bytes.
	var level1 map[string]any
	if err := json.Unmarshal(l.LoadRaw(rawLevel1Data).Data, &level1); err != nil {
		panic(err)
	}
	fmt.Println(level1["name"]) // Prints "level1"

	// Now let's try using audio resources.
	// Audio resources wrap the sound into an *audio.Player
	// that is ready to be used. Every AudioID has its own audio player.
	// Most of the time, if you want to play a sound, you need
	// to rewind the player before doing that.
	a := l.LoadWAV(audioExample)
	if err := a.Player.Rewind(); err != nil {
		panic(err)
	}
	a.Player.Play()

	// Output:
	// level1
}

// This is our stub for the real data.
// In reality, you would probably use a combination of
// go:embed store and real filesystem.
var resdata = map[string][]byte{
	"maps/level1.json": []byte(`{"name": "level1"}`),
	"maps/level2.json": []byte(`{"name": "level2"}`),
	"config.txt":       []byte("some example config\n"),

	// Some minimal-size valid wav resource.
	"audio/example.wav": []byte(strings.Join([]string{
		"\x52\x49\x46\x46\x24\x00\x00\x00\x57\x41\x56\x45\x66\x6d\x74",
		"\x20\x10\x00\x00\x00\x01\x00\x01\x00\x44\xac\x00\x00\x88\x58",
		"\x01\x00\x02\x00\x10\x00\x64\x61\x74\x61\x00\x00\x00\x00",
	}, "")),
}
