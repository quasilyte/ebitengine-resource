package resource

// registry is a resource metadata association index.
//
// Right now it's implemented as a map, but it could become a
// mixture of a slice+map, since most games and assets
// use 0-N keys defined with iota, so it could be better
// to store them inside a slice storage.
//
// We use an opaque type here to make it an implementation detail.
// The users have only Set and Assign operations.
type registry[IDType ~int, InfoType any] struct {
	mapping map[IDType]InfoType
}

// Set binds the typed resource ID to its metadata.
// After that, a respective Load can properly load the resource
// by that ID if the metadata is valid (for instance, the resource path).
//
// If id was unbound before, it will be bound.
// If id was bound before, its metadata will be replaced.
//
// The typed ID could be of type:
// AudioID, FontID, ImageID, RawID, ShaderID.
// The metadata should have a respective type too:
// AudioInfo, FontInfo, ImageInfo, RawInfo, ShaderInfo.
func (r *registry[IDType, InfoType]) Set(id IDType, info InfoType) {
	r.mapping[id] = info
}

// Assign is a convenience wrapper over Set to bind multiple key-value
// pairs at once. See Set documentation for more info.
func (r *registry[IDType, InfoType]) Assign(m map[IDType]InfoType) {
	for k, v := range m {
		r.Set(k, v)
	}
}
