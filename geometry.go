package main

import (
	"math"
)

var (
	xAxis = Vector3D{X: 1, Y: 0, Z: 0}
	yAxis = Vector3D{X: 0, Y: 1, Z: 0}
	zAxis = Vector3D{X: 0, Y: 0, Z: 1}
)

const (
	degToRad = math.Pi / 180
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

type Mat4 [4][4]float64

func dot(row1, row2 [4]float64) float64 {
	return row1[0]*row2[0] + row1[1]*row2[1] + row1[2]*row2[2] + row1[3]*row2[3]
}

func (pt Point3D) Transform(m Mat4) (ret Point3D) {
	ptRow := [4]float64{pt.X, pt.Y, pt.Z, 1}
	ret.X = dot(m[0], ptRow)
	ret.Y = dot(m[1], ptRow)
	ret.Z = dot(m[2], ptRow)
	return
}

func (m Mat4) Mult(m2 Mat4) (ret Mat4) {
	ret[0][0] = dot(m[0], [4]float64{m2[0][0], m2[1][0], m2[2][0], m2[3][0]})
	ret[0][1] = dot(m[0], [4]float64{m2[0][1], m2[1][1], m2[2][1], m2[3][1]})
	ret[0][2] = dot(m[0], [4]float64{m2[0][2], m2[1][2], m2[2][2], m2[3][2]})
	ret[0][3] = dot(m[0], [4]float64{m2[0][3], m2[1][3], m2[2][3], m2[3][3]})

	ret[1][0] = dot(m[1], [4]float64{m2[0][0], m2[1][0], m2[2][0], m2[3][0]})
	ret[1][1] = dot(m[1], [4]float64{m2[0][1], m2[1][1], m2[2][1], m2[3][1]})
	ret[1][2] = dot(m[1], [4]float64{m2[0][2], m2[1][2], m2[2][2], m2[3][2]})
	ret[1][3] = dot(m[1], [4]float64{m2[0][3], m2[1][3], m2[2][3], m2[3][3]})

	ret[2][0] = dot(m[2], [4]float64{m2[0][0], m2[1][0], m2[2][0], m2[3][0]})
	ret[2][1] = dot(m[2], [4]float64{m2[0][1], m2[1][1], m2[2][1], m2[3][1]})
	ret[2][2] = dot(m[2], [4]float64{m2[0][2], m2[1][2], m2[2][2], m2[3][2]})
	ret[2][3] = dot(m[2], [4]float64{m2[0][3], m2[1][3], m2[2][3], m2[3][3]})

	ret[3][0] = dot(m[3], [4]float64{m2[0][0], m2[1][0], m2[2][0], m2[3][0]})
	ret[3][1] = dot(m[3], [4]float64{m2[0][1], m2[1][1], m2[2][1], m2[3][1]})
	ret[3][2] = dot(m[3], [4]float64{m2[0][2], m2[1][2], m2[2][2], m2[3][2]})
	ret[3][3] = dot(m[3], [4]float64{m2[0][3], m2[1][3], m2[2][3], m2[3][3]})
	return
}

func CreateIdentity() (ret Mat4) {
	ret[0][0] = 1
	ret[1][1] = 1
	ret[2][2] = 1
	ret[3][3] = 1
	return
}

func CreateTranslate(x, y, z float64) Mat4 {
	ret := CreateIdentity()
	ret[0][3] = x
	ret[1][3] = y
	ret[2][3] = z
	return ret
}

func CreateScale(x, y, z float64) (ret Mat4) {
	ret[0][0] = x
	ret[1][1] = y
	ret[2][2] = z
	ret[3][3] = 1
	return
}

func CreateRotationX(degrees float64) Mat4 {
	ret := CreateIdentity()
	radAngle := degrees * degToRad

	ret[1][1] = math.Cos(radAngle)
	ret[1][2] = -math.Sin(radAngle)

	ret[2][1] = math.Sin(radAngle)
	ret[2][2] = math.Cos(radAngle)

	return ret
}

func CreateRotationY(degrees float64) Mat4 {
	ret := CreateIdentity()
	radAngle := degrees * degToRad

	ret[0][0] = math.Cos(radAngle)
	ret[0][2] = math.Sin(radAngle)

	ret[2][0] = -math.Sin(radAngle)
	ret[2][2] = math.Cos(radAngle)

	return ret
}

func CreateRotationZ(degrees float64) Mat4 {
	ret := CreateIdentity()
	radAngle := degrees * degToRad

	ret[0][0] = math.Cos(radAngle)
	ret[0][1] = -math.Sin(radAngle)

	ret[1][0] = math.Sin(radAngle)
	ret[1][1] = math.Cos(radAngle)

	return ret
}

func (m Mat4) Inverse() Mat4 {
	// Inverse!
	return m
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

/*
func (s Sphere) Hit(r Ray) (count uint8, t1, t2 float64) {
	rToS := r.Origin.Sub(s.Center)
	A := r.Direction.Dot(r.Direction)
	B := 2 * rToS.Dot(r.Direction)
	C := rToS.Dot(rToS) - math.Pow(s.Radius, 2)
	dtmt := math.Pow(B, 2) - 4*A*C
	if dtmt < 0 {
		return
	}
	divisor := 2 * A

	if dtmt != 0 {
		sqrt := math.Sqrt(dtmt)
		t1 = (-B + sqrt) / divisor
		t2 = (-B - sqrt) / divisor
		if t2 < t1 {
			temp := t1
			t1 = t2
			t2 = temp
		}
		if t1 < 0 {
			count = 0
		} else {
			count = 2
		}
	} else {
		count = 1
		t1 = -B / divisor
	}

	return
}
*/

/*
func (s Sphere) Color(light Light, pt, eye Point3D) image.Color {
	normal := pt.Sub(s.Center).Normalize()
	view := eye.Sub(pt).Normalize()
	L := light.Sub(pt).Normalize()
	diffuse := light.LColor.Mult(s.Mat.Diffuse).Scale(math.Min(1.0, math.Max(0.0,
		normal.Dot(L))))
	specular := light.LColor.Mult(s.Mat.Specular).Scale(math.Pow(math.Min(1.0,
		math.Max(0.0, normal.Dot(L.Add(view).Normalize()))), 128*s.Mat.Shininess))
	ambient := light.LColor.Mult(s.Mat.Ambient)
	return diffuse.Add(specular).Add(ambient)
}
*/

/*
func (s Sphere) Material() Material {
	return s.Mat
}
*/
