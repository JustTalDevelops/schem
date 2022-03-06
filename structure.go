package schem

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"reflect"
)

// schematic implements the structure of a Schematic, providing methods to read from it.
type schematic struct {
	Data    map[string]interface{}
	w, h, l int

	palette []string
	blocks  []uint16
}

// init initialises the schematic structure, parsing several values from the NBT data.
func (s *schematic) init() error {
	s.w, s.h, s.l = int(s.Data["Width"].(int16)), int(s.Data["Height"].(int16)), int(s.Data["Length"].(int16))
	s.blocks = make([]uint16, s.w*s.h*s.l)

	paletteV, dataV := reflect.ValueOf(s.Data["Palette"]), reflect.ValueOf(s.Data["BlockData"])
	paletteSlice, dataSlice := reflect.MakeSlice(reflect.SliceOf(paletteV.Type().Elem()), paletteV.Len(), paletteV.Len()),
		reflect.MakeSlice(reflect.SliceOf(dataV.Type().Elem()), dataV.Len(), dataV.Len())

	reflect.Copy(paletteSlice, paletteV)
	reflect.Copy(dataSlice, dataV)

	s.palette = paletteSlice.Interface().([]string)

	data := dataSlice.Interface().([]byte)
	index, i := 0, 0
	for i < len(data) {
		value, varIntLen := 0, 0
		for true {
			varIntLen++
			value |= (data[i] & 127) << (varIntLen * 7)
			if varIntLen > 5 {
				return fmt.Errorf("varint too long")
			}
			if (data[i] & 128) != 128 {
				i++
				break
			}
			i++
		}

		y := index / (s.w * s.l)
		z := (index % (s.w * s.l)) / s.w
		x := (index % (s.w * s.l)) % s.w

		s.blocks[(y*s.l+z)*s.w+x] = uint16(value)
		index++
	}
	return nil
}

// Dimensions returns the dimensions of the schematic as width, height and length respectively.
func (s *schematic) Dimensions() [3]int {
	return [3]int{s.w, s.h, s.l}
}

// At returns the block found at a given position in the schematic. If any of the X, Y or Z coordinates passed
// are out of the bounds of the schematic, At will panic.
func (s *schematic) At(x, y, z int, _ func(int, int, int) world.Block) (world.Block, world.Liquid) {
	index := (y*s.l+z)*s.w + x
	state := s.palette[s.blocks[index]]
	if state == "minecraft:air" {
		// Don't write air blocks: We simply return nil so that no block is placed at all.
		return nil, nil
	}

	converted, ok := editionConversion[state]
	if !ok {
		// Something went wrong, so we can just treat this as air.
		return nil, nil
	}

	ret, ok := world.BlockByName(converted.Name, converted.Properties)
	if !ok {
		return block.Air{}, nil
	}
	return ret, nil
}
