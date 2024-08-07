package video

import "github.com/veandco/go-sdl2/sdl"

type VideoBackend struct {
    
    sdlWindow *sdl.Window
    sdlRenderer *sdl.Renderer
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

    backend.sdlRenderer.SetDrawColor(0x1f, 0x1f, 0x1f, 0xff)
    backend.sdlRenderer.Clear()

    backend.sdlRenderer.Present()

	return nil
}
