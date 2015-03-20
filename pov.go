package main

import (
	"bufio"
	"errors"
	"github.com/go-gl/mathgl/mgl64"
	"io"
	"math"
	"strconv"
	"unicode"
)

var (
	eye = camera{location: Point3D{X: 0, Y: 0, Z: 0},
		up:     Vector3D{X: 0, Y: 1, Z: 0},
		right:  Vector3D{X: 1.333, Y: 0, Z: 0},
		lookAt: Point3D{X: 0, Y: 0, Z: -1}}

	objects = make([]castable, 0, 10)
	lights  = make([]light, 0, 1)

	eofErr = errors.New("Unexpected EOF")
)

const (
	degToRad = math.Pi / 180
)

type castable interface {
	Hit(r Ray) (bool, float64)
	Color() fColor
	Normal(pt Point3D) Vector3D
	Finish() finish
}

type object struct {
	transforms mgl64.Mat4
	pigment    fColor
	finish     finish
}

type fColor struct {
	R, G, B, A float64
}

type finish struct {
	ambient, diffuse, specular, roughness float64
	reflection, refraction, ior           float64
}

type camera struct {
	location  Point3D
	up, right Vector3D
	lookAt    Point3D
}

type light struct {
	location Point3D
	color    fColor
}

type box struct {
	corner1, corner2 Point3D
	object
}

type sphere struct {
	center Point3D
	radius float64
	object
}

type cone struct {
	end1, end2       Point3D
	radius1, radius2 float64
	object
}

type plane struct {
	normal   Vector3D
	distance float64
	object
}

type triangle struct {
	corner1, corner2, corner3 Point3D
	object
}

type errScanner struct {
	scanner *bufio.Scanner
	err     error
}

type errFloatConv struct {
	err error
}

func (c fColor) RGBA() (r, g, b, a uint32) {
	return uint32(math.Min(c.R*c.A*math.MaxUint16, math.MaxUint16)),
		uint32(math.Min(c.G*c.A*math.MaxUint16, math.MaxUint16)),
		uint32(math.Min(c.B*c.A*math.MaxUint16, math.MaxUint16)),
		uint32(math.Min(c.A*math.MaxUint16, math.MaxUint16))
}

func (c fColor) rgba() (r, g, b, a float64) {
	return c.A * c.R, c.A * c.G, c.A * c.B, c.A
}

func (c fColor) Add(clr fColor) fColor {
	r, g, b, a := clr.rgba()
	fr, fg, fb, fa := c.rgba()
	return fColor{R: fr + r,
		G: fg + g,
		B: fb + b,
		A: math.Min(1.0, fa+a)}
}

func (c fColor) Mult(clr fColor) fColor {
	r, g, b, a := clr.rgba()
	fr, fg, fb, fa := c.rgba()
	return fColor{R: fr * r,
		G: fg * g,
		B: fb * b,
		A: math.Min(1.0, fa*a)}
}

func (c fColor) Scale(factor float64) fColor {
	return fColor{R: c.R * factor,
		G: c.G * factor,
		B: c.B * factor,
		A: c.A}
}

func (es *errScanner) Text() string {
	if es.err != nil {
		return ""
	}
	if !es.scanner.Scan() {
		es.err = eofErr
		return ""
	}
	text := es.scanner.Text()
	return text
}

func (efc *errFloatConv) convert(s string) float64 {
	if efc.err != nil {
		return 0
	}
	var ret float64
	ret, efc.err = strconv.ParseFloat(s, 64)
	return ret
}

func (obj *object) init() {
	obj.transforms = mgl64.Ident4()
	obj.finish.ambient = 0.1
	obj.finish.diffuse = 0.6
	obj.finish.specular = 0.0
	obj.finish.roughness = 0.05
	obj.finish.reflection = 0.0
	obj.finish.refraction = 0.0
	obj.finish.ior = 1.0
}

func makeBox() (b box) {
	b.init()
	return
}

func makeSphere() (s sphere) {
	s.init()
	return
}

func makeCone() (c cone) {
	c.init()
	return
}

func makePlane() (p plane) {
	p.init()
	return
}

func makeTriangle() (t triangle) {
	t.init()
	return
}

func parsePOV(reader io.Reader) (err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(scanPOV)
	for scanner.Scan() {
		switch scanner.Text() {
		case "camera":
			err = parseCamera(scanner)
		case "light_source":
			err = parseLight(scanner)
		case "box":
			err = parseBox(scanner)
		case "sphere":
			err = parseSphere(scanner)
		case "cone":
			err = parseCone(scanner)
		case "plane":
			err = parsePlane(scanner)
		case "triangle":
			err = parseTriangle(scanner)
		default:
			token := scanner.Text()
			if len(token) > 1 && token[:2] == "//" {
				scanner.Split(bufio.ScanLines)
				scanner.Scan()
				scanner.Split(scanPOV)
			}
			// Ignore Unexpected
		}
		if err != nil {
			return
		}
	}
	return
}

func skipBlock(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		switch scanner.Text() {
		case "}":
			return nil
		case "{":
			skipBlock(scanner)
		}
	}
	return eofErr
}

func parseCamera(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}

	var err error
	for scanner.Scan() {
		token := scanner.Text()
		switch token {
		case "location":
			eye.location, err = parsePoint(scanner)
			if err != nil {
				return err
			}
		case "up":
			eye.up, err = parseVector(scanner)
			if err != nil {
				return err
			}
		case "right":
			eye.right, err = parseVector(scanner)
			if err != nil {
				return err
			}
		case "look_at":
			eye.lookAt, err = parsePoint(scanner)
			if err != nil {
				return err
			}
		case "}":
			return nil
		default:
			return errors.New("Unexpected token: '" + token + "'")
		}
	}
	return eofErr
}

func parseLight(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}

	l := light{}
	var err error
	l.location, err = parsePoint(scanner)
	if err != nil {
		return err
	}
	// check for 'color rgb' between vectors
	if !scanner.Scan() || scanner.Text() != "color" ||
		!scanner.Scan() || scanner.Text() != "rgb" {
		return errors.New("error parsing light")
	}
	temp, err := parseVector(scanner)
	if err != nil {
		return err
	}
	l.color = fColor{R: temp.X, G: temp.Y, B: temp.Z, A: 1.0}

	lights = append(lights, l)

	return nil
}

func parseBox(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}
	return skipBlock(scanner)
}

func parseSphere(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}
	s := makeSphere()
	var err error
	s.center, err = parsePoint(scanner)
	if err != nil {
		return err
	}
	if !scanner.Scan() {
		return eofErr
	}
	s.radius, err = strconv.ParseFloat(scanner.Text(), 64)
	err = s.finishObject(scanner)
	if err == nil {
		objects = append(objects, s)
	}

	return err
}

func parseCone(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}
	return skipBlock(scanner)
}

func parsePlane(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}
	p := makePlane()
	var err error
	p.normal, err = parseVector(scanner)
	if err != nil {
		return err
	}
	if !scanner.Scan() {
		return eofErr
	}
	p.distance, err = strconv.ParseFloat(scanner.Text(), 64)
	err = p.finishObject(scanner)
	if err == nil {
		objects = append(objects, p)
	}

	return err
}

func parseTriangle(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}
	return skipBlock(scanner)
}

func parseFinish(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}
	return skipBlock(scanner)
}

func parsePoint(scanner *bufio.Scanner) (Point3D, error) {
	pt := Point3D{}
	es := errScanner{scanner: scanner, err: nil}
	text := es.Text()
	if text != "<" {
		return pt, errors.New("Expected vector, found: '" + text + "'")
	}

	efc := errFloatConv{}

	pt.X = efc.convert(es.Text())
	pt.Y = efc.convert(es.Text())
	pt.Z = efc.convert(es.Text())

	if efc.err != nil {
		return pt, efc.err
	}

	if es.Text() != ">" {
		return pt, errors.New("Unterminated Point")
	}

	if es.err != nil {
		return pt, es.err
	}

	return pt, nil
}

func parseVector(scanner *bufio.Scanner) (Vector3D, error) {
	pt, err := parsePoint(scanner)
	return pt.AsVector(), err
}

func parseScale(scanner *bufio.Scanner) (error, Vector3D) {
	vec := Vector3D{}
	es := errScanner{scanner: scanner, err: nil}
	efc := errFloatConv{}
	text := es.Text()
	if text != "<" {
		scale := efc.convert(text)
		if efc.err != nil {
			return efc.err, vec
		}
		return nil, Vector3D{X: scale, Y: scale, Z: scale}
	}

	vec.X = efc.convert(es.Text())
	vec.Y = efc.convert(es.Text())
	vec.Z = efc.convert(es.Text())

	if efc.err != nil {
		return efc.err, vec
	}
	if es.Text() != ">" {
		return errors.New("Unterminated Vector"), vec
	}
	if es.err != nil {
		return es.err, vec
	}

	return nil, vec
}

func parsePigment(scanner *bufio.Scanner) (fColor, error) {
	// check and skip for { color rgb[f]
	if !scanner.Scan() || scanner.Text() != "{" ||
		!scanner.Scan() || scanner.Text() != "color" ||
		!scanner.Scan() || (scanner.Text() != "rgb" && scanner.Text() != "rgbf") {
		return fColor{}, errors.New("Invalid pigment structure")
	}
	c, err := parseColor(scanner)
	if err == nil && (!scanner.Scan() || scanner.Text() != "}") {
		err = errors.New("Invalid pigment structure")
	}
	return c, err
}

func parseColor(scanner *bufio.Scanner) (fColor, error) {
	c := fColor{}
	es := errScanner{scanner: scanner, err: nil}
	text := es.Text()

	if text != "<" {
		return c, errors.New("Expected color, found: '" + text + "'")
	}

	efc := errFloatConv{}

	c.R = efc.convert(es.Text())
	c.G = efc.convert(es.Text())
	c.B = efc.convert(es.Text())
	c.A = 1.0

	nextTok := es.Text()
	if nextTok != ">" {
		c.A = 1 - efc.convert(nextTok)
		nextTok = es.Text()
	}

	if efc.err != nil {
		return c, efc.err
	}

	if nextTok != ">" {
		return c, errors.New("Unterminated Color")
	}

	if es.err != nil {
		return c, es.err
	}

	return c, nil
}

// Scanner split function to parse pov vector, calling scan will scan a single
// '>', '<', or a whole word up until whitespace or a comma
func scanPOV(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Copied & Modified from bufio.ScanWords
	start := 0
	// Skip leading space, ','
	shouldSkip := func(c byte) bool {
		return unicode.IsSpace(rune(c)) || c == ','
	}

	isToken := func(c byte) bool {
		return shouldSkip(c) || c == '>' || c == '<' || c == '{' || c == '}'
	}

	for ; start < len(data); start++ {
		c := data[start]
		if !shouldSkip(c) {
			break
		}
	}

	if start < len(data) && isToken(data[start]) {
		return start + 1, data[start : start+1], nil
	}

	// Scan until token
	for i := start; i < len(data); i++ {
		c := data[i]
		if isToken(c) || c == '>' || c == '<' {
			return i, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final word, return it
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// get more
	return start, nil, nil
}

func (obj *object) finishObject(scanner *bufio.Scanner) error {
	var err error
	// Ignore transformations for now
	//	var vec Vector3D
	for scanner.Scan() {
		switch scanner.Text() {
		case "translate":
			_, err = parseVector(scanner)
			//			obj.transforms = mgl64.Translate3D(vec.X, vec.Y, vec.Z).Mul4(obj.transforms)
		case "rotate":
			_, err = parseVector(scanner)
			//			obj.transforms = mgl64.HomogRotate3DZ(degToRad * vec.Z).Mul4(
			//				mgl64.HomogRotate3DY(degToRad * vec.Y)).Mul4(
			//				mgl64.HomogRotate3DX(degToRad * vec.X)).Mul4(
			//				obj.transforms)
		case "scale":
			_, err = parseVector(scanner)
			//			obj.transforms = mgl64.Scale3D(vec.X, vec.Y, vec.Z).Mul4(obj.transforms)
		case "pigment":
			obj.pigment, err = parsePigment(scanner)
		case "finish":
			err = obj.parseFinish(scanner)
		case "}":
			return nil
		}
		if err != nil {
			return err
		}
	}
	return eofErr
}

func (obj *object) parseFinish(scanner *bufio.Scanner) error {
	if !scanner.Scan() || scanner.Text() != "{" {
		return errors.New("Missing '{' token")
	}

	var err error
	for scanner.Scan() {
		token := scanner.Text()
		if token == "}" {
			return nil
		}
		if !scanner.Scan() {
			return eofErr
		}
		switch token {
		case "ambient":
			obj.finish.ambient, err = strconv.ParseFloat(scanner.Text(), 64)
		case "diffuse":
			obj.finish.diffuse, err = strconv.ParseFloat(scanner.Text(), 64)
		case "specular":
			obj.finish.specular, err = strconv.ParseFloat(scanner.Text(), 64)
		case "roughness":
			obj.finish.roughness, err = strconv.ParseFloat(scanner.Text(), 64)
		case "reflection":
			obj.finish.reflection, err = strconv.ParseFloat(scanner.Text(), 64)
		case "refraction":
			obj.finish.refraction, err = strconv.ParseFloat(scanner.Text(), 64)
		case "ior":
			obj.finish.ior, err = strconv.ParseFloat(scanner.Text(), 64)
		default:
			return errors.New("Unexpected token: '" + token + "'")
		}
		if err != nil {
			return err
		}
	}
	return eofErr
}

func (obj object) transform(pt Point3D) Point3D {
	return pt.Transform(obj.transforms)
}

func (s sphere) Hit(r Ray) (hitObj bool, t1 float64) {
	/* get inverse sphere transform matrix, then apply to ray */
	invM := s.transforms.Inv()
	transRay := Ray{Origin: r.Origin.Transform(invM), Direction: r.Direction.Transform(invM).Normalize()}
	rToS := transRay.Origin.Sub(s.transform(s.center))
	A := transRay.Direction.Dot(transRay.Direction)
	B := 2 * rToS.Dot(transRay.Direction)
	C := rToS.Dot(rToS) - math.Pow(s.radius, 2)
	dtmt := math.Pow(B, 2) - 4*A*C
	if dtmt < 0 {
		return
	}

	divisor := 2 * A

	if dtmt != 0 {
		sqrt := math.Sqrt(dtmt)
		t1 = (-B + sqrt) / divisor
		t2 := (-B - sqrt) / divisor
		if t2 < t1 {
			t1 = t2
		}
		if t1 < 0 {
			hitObj = false
		} else {
			hitObj = true
		}
	} else {
		hitObj = true
		t1 = -B / divisor
	}
	return
}

func (s sphere) Normal(pt Point3D) Vector3D {
	return pt.Sub(s.transform(s.center)).Normalize()
}

func (p plane) Hit(r Ray) (hitObj bool, t1 float64) {
	vDotN := r.Direction.Dot(p.normal)
	if vDotN != 0 {
		t1 = -(r.Origin.AsVector().Dot(p.normal) - p.distance) / vDotN
		if t1 > 0 {
			hitObj = true
		}
	}
	return
}

func (p plane) Normal(pt Point3D) Vector3D {
	return p.normal
}

func (obj object) Finish() finish {
	return obj.finish
}

func (obj object) Color() fColor {
	return obj.pigment
}
