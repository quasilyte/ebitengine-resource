## Ebitengine Resource Manager Library

![Build Status](https://github.com/quasilyte/ebitengine-resource/workflows/Go/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/quasilyte/ebitengine-resource)](https://pkg.go.dev/mod/github.com/quasilyte/ebitengine-resource)

### Overview

A resource manager (loader) for [Ebitengine](https://github.com/hajimehoshi/ebiten).

**Key features:**

* Resource caching (only the first load decodes the resource)
* Easy to use and opinionated
* iota-friendly typed constants API
* Int-based keys are also efficient, so the lookups are very fast

Some games that were built with this library:

* [Decipherism](https://quasilyte.itch.io/decipherism)
* [Retrowave City](https://quasilyte.itch.io/retrowave-city)
* [Autotanks](https://quasilyte.itch.io/autotanks)

### Installation

```bash
go get github.com/quasilyte/ebitengine-resource
```

### Quick Start

```go
package main

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

func main() {
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
```

### Introduction

How to use this library properly?

You start by creating a loader with [resource.NewLoader()](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#NewLoader). It should happen after you acquired an `*audio.Context` from Ebitengine. It's not recommended to make loader global, pass it as an explicit dependency everywhere you need to access the game resources.

Then you bind the resources using typed registries. For instance, binding image resources is done via `loader.ImageRegistry` field (use `Set` or `Assign` methods).

The loader acts as a cached resource access point. Resources are keyed by their ID. The ID is a simple integer. All metadata is associated with that ID too. It's recommended to make the core resources iota-style constants.

If you want to preload a resource, do a respective `Load` (e.g. [LoadImage](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Loader.LoadImage), [LoadAudio](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Loader.LoadAudio)) call either during a game launch or during the loading screen.

Most types in a package can be described by these categories:

1. ID types that belong to specific kind of resource (e.g. [ImageID](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#ImageID), [AudioID](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#AudioID))
2. Info objects that describe the resource (e.g. [ImageInfo](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#ImageInfo), [AudioInfo](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#AudioInfo))
3. The actual resource objects (e.g. [Image](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Image), [Audio](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Audio))

The info objects should be bound before the resource is accessed via `Load` method. It's possible to bind extra resources during the run-time.

Supported resource kinds:

* [Audio](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Audio) (`*Audio.Player` with decoded stream)
* [Font](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Font) (`font.Face` with relevant properties like font size and line spacing)
* [Image](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Image) (`*ebiten.Image` created from a texture)
* [Shader](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Shader) (a compiled `*ebiten.Shader`)
* [Raw](https://pkg.go.dev/github.com/quasilyte/ebitengine-resource#Raw) (stored as `[]byte`)
