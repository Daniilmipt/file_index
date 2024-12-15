package types

const (
	ChunkSize   = 4096
	HashSize    = 32
	SampleRate  = 0.1
	MaxDistance = 256.0
)

type IndexConfig struct {
	ChunkSize      int
	SampleRate     float64
	ErrorThreshold float64
}
