package video

/*
 * Glyphs are always 8x8
 */
type Glyph [8]byte

type Font struct {
	backend *VideoBackend
	glyphs  []Glyph
}

func (self *Font) getGlyph(index byte) Glyph {
	return self.glyphs[index]
}

func NewFont(backend *VideoBackend) Font {
	var ret Font

	ret.backend = backend
	ret.glyphs = make([]Glyph, 0)

	// Fill the glyphs array with empty shit

	for _, glyph_data := range default_glyph_list {
		ret.glyphs = append(ret.glyphs, glyph_data)
	}

	return ret
}

func (self *Font) DrawGlyph(x int, y int, g byte, color Color, backcolor Color, nesglyph bool) {

	// Grab the pixel draw function from the backend

	var px func(a int32, b int32, clr Color) = self.backend.DrawPixel

	// Check if we want to display a NES relative glyph

	if nesglyph {
		px = self.backend.DrawNESPixel
	}

	// Grab the glyph we want

	glyph := self.getGlyph(g)

	// Loop over the bytes in this glyph

	for idx, d := range glyph {

		// Now loop over the bits inside this byte

		for bit := range 8 {

			// Check if this bit is set

			if (d & (1 << bit)) == (1 << bit) {
				px(int32(x+bit), int32(y+idx), color)
			} else {
				px(int32(x+bit), int32(y+idx), backcolor)
			}
		}
	}
}
