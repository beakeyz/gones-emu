package video

import "github.com/veandco/go-sdl2/sdl"

type VideoBackend struct {
	sdlWindow   *sdl.Window
	sdlRenderer *sdl.Renderer

	defaultFont Font

	deferFlush bool
}

type Color struct {
	r uint8
	g uint8
	b uint8
	a uint8
}

const (
	SCREEN_HEIGHT     = 1000
	SCREEN_WIDTH      = 1150
	NES_SCREEN_HEIGHT = 240
	NES_SCREEN_WIDTH  = 256

	// Offset the NES screen a bit from the top left

	NES_SCREEN_X_START = 25
	NES_SCREEN_Y_START = 25

	/* NES pixel to HOST pixel ratio */
	NES_PTHP_RATIO = 2
)

func NewColor(r uint8, g uint8, b uint8, a uint8) Color {
	return Color{r, g, b, a}
}

func InitVideo(backend *VideoBackend) error {
	var err error

	/* Initialize the SDL library for video stuff */
	err = sdl.Init(sdl.INIT_VIDEO)

	if err != nil {
		return err
	}

	/* Initialize a window and a renderer for us to draw with */
	backend.sdlWindow, backend.sdlRenderer, err = sdl.CreateWindowAndRenderer(SCREEN_WIDTH, SCREEN_HEIGHT, 0)

	if err != nil {
		return err
	}

	// Create a new font for us to use

	backend.defaultFont = NewFont(backend)

	// Defer flushing calls

	backend.deferFlush = true

	backend.DrawRect(0, 0, SCREEN_WIDTH, SCREEN_HEIGHT, NewColor(0x1f, 0x1f, 0x1f, 0xff))

	backend.DrawRect(NES_SCREEN_X_START, NES_SCREEN_Y_START, NES_SCREEN_WIDTH*NES_PTHP_RATIO, NES_SCREEN_HEIGHT*NES_PTHP_RATIO, NewColor(0x00, 0x00, 0x00, 0xff))

	backend.DrawText(0, 0, "Hello There! !@#$%^&*()_", NewColor(0xff, 0xff, 0xff, 0xff))

	backend.Flush()

	return nil
}

/*
 * TODO: Rename the 'video' package lmao
 *
 * It's much more than video atm
 */
func (back *VideoBackend) CollectEvent() sdl.Event {
	return sdl.PollEvent()
}

func (back *VideoBackend) IsKeyPressed(a sdl.Keycode) bool {
	sc := sdl.GetScancodeFromKey(a)
	return (sdl.GetKeyboardState()[sc] != 0)
}

func (back *VideoBackend) DrawNESPixel(x int32, y int32, clr Color) {

	if x >= NES_SCREEN_WIDTH || y >= NES_SCREEN_HEIGHT {
		return
	}

	// Add the offset of the NES screen to these coords

	back.DrawRect(x*NES_PTHP_RATIO+NES_SCREEN_X_START, y*NES_PTHP_RATIO+NES_SCREEN_Y_START, NES_PTHP_RATIO, NES_PTHP_RATIO, clr)
}

func (back *VideoBackend) DrawPixel(x int32, y int32, clr Color) {
	// Set the color
	back.sdlRenderer.SetDrawColor(clr.r, clr.g, clr.b, clr.a)

	// Draw the pixel
	back.sdlRenderer.DrawPoint(x, y)
}

func (back *VideoBackend) DrawRect(x int32, y int32, w int32, h int32, clr Color) {
	rect := sdl.Rect{
		X: x,
		Y: y,
		W: w,
		H: h,
	}

	// Set the color
	back.sdlRenderer.SetDrawColor(clr.r, clr.g, clr.b, clr.a)

	// Draw the rect
	back.sdlRenderer.FillRect(&rect)
}

func (back *VideoBackend) SetDeferFlush(def bool) {
	back.deferFlush = def
}

func (back *VideoBackend) Flush() {
	back.sdlRenderer.Present()
}

func (back *VideoBackend) drawText(x int32, y int32, text string, color Color, nes bool) {

	// Loop over all runes inside the provided text

	for _, d := range text {

		// Draw the current glyph

		back.defaultFont.DrawGlyph(int(x), int(y), byte(d), color, nes)

		// Add an offset to the x-coordinate

		x += 8
	}
}

func (back *VideoBackend) DrawNESText(x int32, y int32, text string, color Color) {
	back.drawText(x, y, text, color, true)
}

func (back *VideoBackend) DrawText(x int32, y int32, text string, color Color) {
	back.drawText(x, y, text, color, false)
}
