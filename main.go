// Voronoi interactive display
// 2D view of voronoi cells. Users may click to place points which will re-generate the diagram.
// Modifying the window size will also re-generate the diagram with the new bounding box size.
package main

import (
	"flag"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/goxjs/gl"
	"github.com/goxjs/glfw"
	"github.com/omustardo/gome"
	"github.com/omustardo/gome/camera"
	"github.com/omustardo/gome/core/entity"
	"github.com/omustardo/gome/input/keyboard"
	"github.com/omustardo/gome/input/mouse"
	"github.com/omustardo/gome/shader"
	"github.com/omustardo/gome/util/fps"
	"github.com/omustardo/gome/view"
	"fmt"
	"github.com/pzsz/voronoi"
	"image/color"
	"math/rand"
	"time"
	"github.com/omustardo/voronoi/state"
)

var (
	windowWidth  = flag.Int("window_width", 1000, "initial window width")
	windowHeight = flag.Int("window_height", 1000, "initial window height")

	screenshotPath = flag.String("screenshot_dir", `C:\Users\Omar\Desktop\screenshots\`, "Folder to save screenshots in. Name is the timestamp of when they are taken.")

	// Explicitly listing the base dir is a hack. It's needed because `go run` produces a binary in a tmp folder so we can't
	// use relative asset paths. More explanation in omustardo\gome\asset\asset.go
	baseDir = flag.String("base_dir", `C:\workspace\Go\src\github.com\omustardo\voronoi`, "All file paths should be specified relative to this root.")
)

func main() {
	flag.Parse()
	terminate := gome.Initialize("Voronoi", *windowWidth, *windowHeight, *baseDir)
	defer terminate()

	shader.Model.SetAmbientLight(&color.NRGBA{255, 255, 255, 0})
	cam := camera.NewTargetCamera(&entity.Entity{}, mgl32.Vec3{0, 0, 1})

	// TODO: Add support for window resizing

	w, h := view.Window.GetSize()
	prevW, prevH := w, h
	diagram := state.New(float64(w), float64(h), randSites(float64(w), float64(h), 30))

	ticker := time.NewTicker(time.Second / 60)
	for !view.Window.ShouldClose() {
		fps.Handler.Update()
		glfw.PollEvents() // Reads window events, like keyboard and mouse input.
		keyboard.Handler.Update()
		mouse.Handler.Update()
		w, h = view.Window.GetSize()
		if prevW != w || prevH != h {
			fmt.Printf("w:%d->%d  h:%d->%d\n", prevW, w, prevH, h)
			diagram.SetDimensions(float64(w), float64(h))
			prevW, prevH = w, h
		}

		ApplyInputs(diagram)

		cam.Update(fps.Handler.DeltaTime())

		// Set up Model-View-Projection Matrix and send it to the shader programs.
		mvMatrix := cam.ModelView()
		pMatrix := cam.ProjectionOrthographic(float32(w), float32(h))
		shader.Model.SetMVPMatrix(pMatrix, mvMatrix)

		// Clear screen, then Draw everything
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// model.RenderXYZAxes()

		diagram.Render()

		// Swaps the buffer that was drawn on to be visible. The visible buffer becomes the one that gets drawn on until it's swapped again.
		view.Window.SwapBuffers()
		<-ticker.C // wait up to 1/60th of a second. This caps framerate to 60 FPS.
	}
}

func randSites(w, h float64, count int) []voronoi.Vertex {
	var sites []voronoi.Vertex
	for i := 0; i < count; i++ {
		sites = append(sites, voronoi.Vertex{rand.Float64()*w - w/2, rand.Float64()*h - h/2})
	}
	return sites
}

func ApplyInputs(d *state.State) {
	w, h := view.Window.GetSize()
	if keyboard.Handler.JustPressed(glfw.KeySpace) {
		//util.SaveScreenshot(w, h, filepath.Join(*screenshotPath, fmt.Sprintf("%d.png", util.GetTimeMillis())))
	}

	if mouse.Handler.LeftPressed() && !mouse.Handler.WasLeftPressed() {
		pos := mouse.Handler.Position()
		worldPos := mgl32.Vec3{pos.X() - float32(w)/2, -pos.Y() + float32(h)/2, 0}
		fmt.Println("mouse click at", mouse.Handler.Position(), " world space:", worldPos)

		d.AddPoint(voronoi.Vertex{
			X: float64(worldPos.X()),
			Y: float64(worldPos.Y())})
	}

	// TODO: Maintain list of all points so you can undo placement.
	// Also be able to clear the entire image.
	// Also be able to take a screenshot on webgl.
}
