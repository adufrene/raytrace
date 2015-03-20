package main

import (
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

var (
	xAxis = Vector3D{X: 1, Y: 0, Z: 0}
	yAxis = Vector3D{X: 0, Y: 1, Z: 0}
	zAxis = Vector3D{X: 0, Y: 0, Z: 1}
)

type Ray struct {
	Origin    Point3D
	Direction Vector3D
}

type Point3D struct {
	X, Y, Z float64
}

type Vector3D struct {
	X, Y, Z float64
}

func dot(row1, row2 [4]float64) float64 {
	return row1[0]*row2[0] + row1[1]*row2[1] + row1[2]*row2[2] + row1[3]*row2[3]
}

func (pt Point3D) Transform(m mgl64.Mat4) (ret Point3D) {
	ptRow := [4]float64{pt.X, pt.Y, pt.Z, 1}
	ret.X = dot([4]float64{m[0*4+0], m[1*4+0], m[2*4+0], m[3*4+0]}, ptRow)
	ret.Y = dot([4]float64{m[0*4+1], m[1*4+1], m[2*4+1], m[3*4+1]}, ptRow)
	ret.Z = dot([4]float64{m[0*4+2], m[1*4+2], m[2*4+2], m[3*4+2]}, ptRow)
	return
}

func (vec Vector3D) Transform(m mgl64.Mat4) (ret Vector3D) {
	vecRow := [4]float64{vec.X, vec.Y, vec.Z, 1}
	ret.X = dot([4]float64{m[0*4+0], m[1*4+0], m[2*4+0], m[3*4+0]}, vecRow)
	ret.Y = dot([4]float64{m[0*4+1], m[1*4+1], m[2*4+1], m[3*4+1]}, vecRow)
	ret.Z = dot([4]float64{m[0*4+2], m[1*4+2], m[2*4+2], m[3*4+2]}, vecRow)
	return
}

func CreateRay(pt1, pt2 Point3D) Ray {
	diffVec := pt2.Sub(pt1)
	return Ray{Origin: pt1, Direction: diffVec.Normalize()}
}

func (r Ray) PointAt(t float64) Point3D {
	return r.Origin.Translate(r.Direction.Scale(t))
}

func (pt Point3D) Add(pt2 Point3D) Vector3D {
	return Vector3D{X: pt.X + pt2.X, Y: pt.Y + pt2.Y, Z: pt.Z + pt2.Z}
}

func (pt Point3D) Sub(pt2 Point3D) Vector3D {
	return Vector3D{X: pt.X - pt2.X, Y: pt.Y - pt2.Y, Z: pt.Z - pt2.Z}
}

func (pt Point3D) Scale(vec Vector3D) Point3D {
	return Point3D{X: pt.X * vec.X, Y: pt.Y * vec.Y, Z: pt.Z * vec.Z}
}

func (p1 Point3D) Dist(p2 Point3D) float64 {
	return p1.Sub(p2).Length()
}

func (vec Vector3D) Normalize() Vector3D {
	return vec.Scale(1 / vec.Length())
}

func (vec Vector3D) Length() float64 {
	return math.Sqrt(vec.Dot(vec))
}

func (vec Vector3D) Dot(vec2 Vector3D) float64 {
	return vec.X*vec2.X + vec.Y*vec2.Y + vec.Z*vec2.Z
}

func (vec Vector3D) Scale(multiplier float64) Vector3D {
	return Vector3D{X: vec.X * multiplier, Y: vec.Y * multiplier, Z: vec.Z * multiplier}
}

func (vec Vector3D) Sub(vec2 Vector3D) Vector3D {
	return Vector3D{X: vec.X - vec2.X, Y: vec.Y - vec2.Y, Z: vec.Z - vec2.Z}
}

func (vec Vector3D) Add(vec2 Vector3D) Vector3D {
	return Vector3D{X: vec.X + vec2.X, Y: vec.Y + vec2.Y, Z: vec.Z + vec2.Z}
}

func (pt Point3D) Translate(vec Vector3D) Point3D {
	return Point3D{X: pt.X + vec.X, Y: pt.Y + vec.Y, Z: pt.Z + vec.Z}
}

func (pt Point3D) AsVector() Vector3D {
	return Vector3D{X: pt.X, Y: pt.Y, Z: pt.Z}
}
