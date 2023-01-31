package resource

import (
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type AudioID int

type AudioInfo struct {
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
	ID     AudioID
	Player *audio.Player
	Group  uint
	Volume float64
}
