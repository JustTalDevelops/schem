package schem

import (
	"compress/gzip"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"io"
	"io/ioutil"
	"sync"
)

// Schematic represents a parsed schematic with blocks in it. It may be used to read blocks from.
type Schematic struct {
	*schematic
}

var (
	mu           sync.Mutex
	decompressor *gzip.Reader

	// Check to ensure that *schematic satisfies the world.Structure interface.
	_ world.Structure = (*schematic)(nil)
)

// FromReader attempts to read a Schematic from an io.Reader passed. If successful, the schematic with all its
// data is returned.
func FromReader(r io.Reader) (Schematic, error) {
	mu.Lock()
	if decompressor == nil {
		reader, err := gzip.NewReader(r)
		if err != nil {
			return Schematic{}, fmt.Errorf("decompress schematic: %w", err)
		}
		decompressor = reader
	} else {
		if err := decompressor.Reset(r); err != nil {
			return Schematic{}, fmt.Errorf("decompress schematic: %w", err)
		}
	}
	b, _ := ioutil.ReadAll(decompressor)
	_ = decompressor.Close()
	mu.Unlock()

	m := make(map[string]interface{})
	if err := nbt.UnmarshalEncoding(b, &m, nbt.BigEndian); err != nil {
		return Schematic{}, fmt.Errorf("parse schematic: %w", err)
	}
	s := &schematic{Data: m}
	return Schematic{schematic: s}, s.init()
}
