package resource

type Raw struct {
	ID RawID

	Data []byte
}

type RawInfo struct {
	Path string
}

type RawID int

type RawRegistry struct {
	mapping map[RawID]RawInfo
}

func (r *RawRegistry) Set(id RawID, info RawInfo) {
	r.mapping[id] = info
}

func (r *RawRegistry) Assign(m map[RawID]RawInfo) {
	for k, v := range m {
		r.Set(k, v)
	}
}
