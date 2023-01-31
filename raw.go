package resource

type RawID int

type RawInfo struct {
	Path string
}

type Raw struct {
	ID RawID

	Data []byte
}
