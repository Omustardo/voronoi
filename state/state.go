package state

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/goxjs/gl"
	"github.com/omustardo/gome/core/entity"
	"github.com/omustardo/gome/model"
	"encoding/binary"
	"github.com/omustardo/bytecoder"
	"github.com/omustardo/gome/model/mesh"
	"github.com/pzsz/voronoi"
	"image/color"
	"fmt"
)

type State struct {
	w, h float64
	points []voronoi.Vertex

	// updated is set to false whenever the state is changed, and to true whenever the points and lines models are
	// up to date with the most recent information. This calculation of models is done only once before rendering
	// which prevents unnecessary interaction with opengl buffers.
	updated bool
	pointsModel, linesModel model.Model
}

func New(w, h float64, points []voronoi.Vertex) *State {
	s := &State{
		w: w,
		h: h,
		points: points,
		updated: false,
	}
	s.update()
	return s
}

//// Points returns the points that make up the state.
//func (s *State) Points() []voronoi.Vertex {
//	return s.points[:] // returns a copy
//}

// SetPoints sets the slice of vertices.
func (s *State) SetPoints(points []voronoi.Vertex) {
	if pointsEq(points, s.points) {
		return
	}
	s.points = points
	s.updated = false
}

func (s *State) AddPoint(p voronoi.Vertex) {
	// don't add an already existing point - it seems to break something. perhaps in the voronoi calculations.
	for i := range s.points {
		if p.X == s.points[i].X && p.Y == s.points[i].Y {
			return
		}
	}

	ps := append(s.points, p)
	s.SetPoints(ps)
}

func (s *State) SetDimensions(w, h float64) {
	if s.w == w && s.h == h {
		return
	}
	s.w = w
	s.h = h
	s.updated = false
}

func (s *State) update() {
	d := diagram(s.w, s.h, s.points)
	s.pointsModel = pointCloudModel(diagramPoints(d))
	s.linesModel = linesModel(diagramLines(d))
	s.updated = true

}

func (s *State) Render() {
	if !s.updated {
		s.update()
	}
	s.pointsModel.Render()
	s.linesModel.Render()
}

func diagram(w, h float64, sites []voronoi.Vertex) *voronoi.Diagram {
	bbox := voronoi.NewBBox(-w/2, w/2, -h/2, h/2)
	// Compute diagram and close cells (add half edges from bounding box)
	return voronoi.ComputeDiagram(sites, bbox, true)
}

// diagramLines returns all of the points making up edges in a voronoi diagram.
// A line consists of two endpoints, so the returned sliced will always be of even length.
func diagramLines(diagram *voronoi.Diagram) []mgl32.Vec3 {
	var vertices []mgl32.Vec3
	for _, edge := range diagram.Edges {
		vertices = append(vertices, vertToVec3(edge.Va), vertToVec3(edge.Vb))
	}
	return vertices
}

// linesModel returns a model containing a mesh made up of all of the provided line segments.
func linesModel(lines []mgl32.Vec3) model.Model {
	if len(lines) == 0 {
		return model.Model{}
	}
	vertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, bytecoder.Vec3(binary.LittleEndian, lines...), gl.STATIC_DRAW)

	m := mesh.NewMesh(vertexBuffer, gl.Buffer{}, gl.Buffer{}, gl.LINES, len(lines), &color.NRGBA{255, 255, 255, 255}, gl.Texture{}, gl.Buffer{})
	return model.Model{
		Mesh:   m,
		Entity: entity.Default(),
	}
}

// diagramPoints returns the centers of each cell.
func diagramPoints(diagram *voronoi.Diagram) []mgl32.Vec3 {
	var vertices []mgl32.Vec3
	for _, cell := range diagram.Cells {
		vertices = append(vertices, mgl32.Vec3{float32(cell.Site.X), float32(cell.Site.Y), 0})
	}
	return vertices
}

// pointCloudModel returns a model containing a mesh made up of the provided points.
func pointCloudModel(points []mgl32.Vec3) model.Model {
	if len(points) == 0 {
		return model.Model{}
	}
	vertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, bytecoder.Vec3(binary.LittleEndian, points...), gl.STATIC_DRAW)

	m := mesh.NewMesh(vertexBuffer, gl.Buffer{}, gl.Buffer{}, gl.POINTS, len(points), &color.NRGBA{255, 255, 255, 255}, gl.Texture{}, gl.Buffer{})
	return model.Model{
		Mesh:   m,
		Entity: entity.Default(),
	}
}

func vertToVec3(v voronoi.EdgeVertex) mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), 0}
}

func pointsEq(p1, p2 []voronoi.Vertex) bool {
	if len(p1) != len(p2) {
		return false
	}
	m := make(map[string]bool)
	for _, p := range p1 {
		m[fmt.Sprintf("%f:%f", p.X, p.Y)] = true
	}
	for _, p := range p2 {
		if !m[fmt.Sprintf("%f:%f", p.X, p.Y)] {
			return false
		}
	}
	return true
}