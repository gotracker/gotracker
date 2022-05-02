package song

// Row is an interface to a row
type Row[TChannelData any] interface {
	GetChannels() []TChannelData
}
