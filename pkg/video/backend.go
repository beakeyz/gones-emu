package video

import "github.com/veandco/go-sdl2/sdl"

type VideoBackend struct {
    sdlWindow *sdl.Window
    sdlRenderer *sdl.Renderer

    deferFlush bool
}

type Color struct {
    r uint8
    g uint8
    b uint8
    a uint8
}

func InitVideo(backend *VideoBackend) error {
    var err error 

    /* Initialize the SDL library for video stuff */
	err = sdl.Init(sdl.INIT_VIDEO)

    if (err != nil) {
        return err
    }

    /* Initialize a window and a renderer for us to draw with */
    backend.sdlWindow, backend.sdlRenderer, err = sdl.CreateWindowAndRenderer(512, 512, 0)

    if (err != nil) {
        return err
    }

    backend.deferFlush = true 

    backend.DrawRect(0, 0, 512, 512, Color { 0x1f, 0x1f, 0x1f, 0xff })

    backend.DrawPixel(0, 0, Color{ 0x00, 0x00, 0xff, 0xff});

    backend.DrawRect(50, 50, 52, 52, Color { 0x1f, 0x00, 0x1f, 0xff })

    backend.Flush()
	return nil
}

func (back *VideoBackend) DrawPixel(x int32, y int32, clr Color) {
    // Set the color
    back.sdlRenderer.SetDrawColor(clr.r, clr.g, clr.b, clr.a);

    // Draw the pixel
    back.sdlRenderer.DrawPoint(x, y)
}

func (back *VideoBackend) DrawRect(x int32, y int32, w int32, h int32, clr Color) {
    rect := sdl.Rect {
        X: x,
        Y: y,
        W: w,
        H: h,
    }

    // Set the color
    back.sdlRenderer.SetDrawColor(clr.r, clr.g, clr.b, clr.a);

    // Draw the rect
    back.sdlRenderer.FillRect(&rect)
}

func (back *VideoBackend) SetDeferFlush(def bool) {
    back.deferFlush = def
}

func (back *VideoBackend) Flush() {
    back.sdlRenderer.Present()
}
