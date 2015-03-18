package main

import (
	"fmt"
	//	"github.com/go-gl/mathgl/mgl64"
	"image"
	"image/jpeg"
	"math"
	"os"
	"strings"
)

var (
	/* For chromebook, save to downloads so file can be viewed from chromeos */
	fileDir = os.Getenv("HOME") + "/Downloads/"
	ext     = ".jpg"

	imgWidth  = 640
	imgHeight = 480
	//	imgWidth  = 20
	//	imgHeight = 14

	red   = fColor{R: 1.0, G: 0.0, B: 0.0, A: 1.0}
	green = fColor{R: 0.0, G: 0.0, B: 0.0, A: 1.0}
	blue  = fColor{R: 0.0, G: 0.0, B: 1.0, A: 1.0}
	white = fColor{R: 1.0, G: 1.0, B: 1.0, A: 1.0}
	black = fColor{R: 0.0, G: 0.0, B: 0.0, A: 1.0}

	bkgndColor = black // fColor{R: 0.2706, G: 0.3137, B: 0.3294, A: 1.0}

	MAX_DEPTH = 7
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage:", os.Args[0], "<path-to-pov-file>")
		return
	}

	filename := args[0]

	povFile, err := os.Open(filename)
	defer povFile.Close()
	if err == nil {
		err = parsePOV(povFile)
	}
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	img := image.NewRGBA(image.Rectangle{image.ZP, image.Point{imgWidth, imgHeight}})

	xTrans := eye.right.Scale(2 / float64(imgWidth))
	yTrans := eye.up.Scale(2 / float64(imgHeight))

	xStart := eye.location.Translate(eye.right.Scale(-1))
	yStart := eye.location.Translate(eye.up.Scale(-1))
	imgPlane := eye.lookAt.Sub(eye.location).Normalize().Scale(2)

	currX := xStart
	for x := 0; x < imgWidth; x++ {
		currY := yStart
		for y := imgHeight - 1; y >= 0; y-- {
			view := eye.location.Translate(currX.Sub(eye.location)).
				Translate(currY.Sub(eye.location)).Translate(imgPlane)
			_, color := castRay(CreateRay(eye.location, view), 0)
			img.Set(x, y, color)
			currY = currY.Translate(yTrans)
		}
		currX = currX.Translate(xTrans)
	}

	splitString := strings.Split(filename, "/")
	name := splitString[len(splitString)-1]
	if strings.HasSuffix(name, ".pov") {
		dotSplit := strings.Split(name, ".")
		name = strings.Join(dotSplit[:len(dotSplit)-1], ".")
	}
	outFile := fileDir + name + ext
	file, err := os.Create(outFile)

	if err != nil {
		panic(err)
	}
	defer func() {
		file.Close()
	}()

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		panic(err)
	}
}

func castRay(ray Ray, depth int) (bool, fColor) {
	if depth > MAX_DEPTH {
		return false, bkgndColor
	}

	if count, t1, _, ndx := hitAnything(ray); count > 0 {
		obj := objects[ndx]
		pxlClr := fColor{}
		origPt := ray.PointAt(t1)
		interPt := ray.PointAt(t1 - 0.01)
		for i := range lights {
			light := lights[i]
			if !isShadowed(interPt, light, ndx) {
				pxlClr = pxlClr.Add(calcColor(obj, light, interPt, eye.location))
				normal := obj.Normal(interPt)
				if obj.Finish().reflection > 0 {
					reflection := ray.Direction.Sub(normal.Scale(2 * ray.Direction.Dot(normal)))
					if reflect, color := castRay(Ray{interPt, reflection.Normalize()}, depth+1); reflect {
						pxlClr = pxlClr.Add(color.Scale(obj.Finish().reflection))
					}
				}
				if obj.Finish().refraction > 0 {
					// Assuming non object material is air w/ ior=1
					var internal bool
					var refractRay Ray
					//					internal, refractRay = calcRefractRay(ray, obj, origPt, 1, 1)
					if ray.Direction.Dot(normal) > 0 { // We are exiting the object
						internal, refractRay = calcRefractRay(ray, obj, origPt, obj.Finish().ior, 1)
					} else { // We are entering the object
						internal, refractRay = calcRefractRay(ray, obj, origPt, 1, obj.Finish().ior)
					}
					if !internal {
						if refract, color := castRay(refractRay, depth+1); refract {
							pxlClr = pxlClr.Add(color.Scale(obj.Finish().refraction))
						}
					}
				}
			} else {
				pxlClr = pxlClr.Add(light.color.Mult(obj.Color().
					Scale(obj.Finish().ambient)))
			}
		}
		return true, pxlClr
	}
	return false, bkgndColor
}

func calcRefractRay(initialRay Ray, obj castable, origPt Point3D,
	n1, n2 float64) (internalReflection bool, refractRay Ray) {
	// (n_1 ( d - n ( d . n)) / n_2) - (n * sqrt( 1 - ( n_1^2 ( 1 - ( d . n)^2) / n_2^2))
	normal := obj.Normal(origPt)
	dDotN := initialRay.Direction.Dot(normal)
	sqrtComp := math.Pow(n1, 2) * (1 - math.Pow(dDotN, 2)) / math.Pow(n2, 2)
	if sqrtComp > 1 {
		return true, Ray{}
	}
	refract := initialRay.Direction.Sub(normal.Scale(
		dDotN)).Scale(n1 / n2).Sub(
		normal.Scale(math.Sqrt(1 - sqrtComp))).Normalize()
	// Make ray start w/in object
	return false, Ray{Origin: origPt.Translate(refract.Scale(0.01)),
		Direction: refract}
}

func isShadowed(pt Point3D, light light, objNdx int) bool {
	r := CreateRay(pt, light.location)
	for ndx := range objects {
		if ndx != objNdx {
			if count, _, _ := objects[ndx].Hit(r); count > 0 {
				return true
			}
		}
	}
	return false
}

func hitAnything(r Ray) (count uint8, t1, t2 float64, hitNdx int) {
	t1 = math.MaxFloat64
	for ndx := range objects {
		if hitCount, hit1, hit2 := objects[ndx].Hit(r); hitCount > 0 && hit1 < t1 {
			count, t1, t2 = hitCount, hit1, hit2
			hitNdx = ndx
		}
	}
	return
}

func calcColor(obj castable, light light, pt, eye Point3D) fColor {
	normal := obj.Normal(pt)
	view := eye.Sub(pt).Normalize()
	L := light.location.Sub(pt).Normalize()
	diffuse := light.color.Mult(obj.Color().Scale(obj.Finish().diffuse)).
		Scale(math.Min(1.0, math.Max(0.0, normal.Dot(L))))
	specular := light.color.Mult(obj.Color().Scale(obj.Finish().specular)).
		Scale(math.Pow(math.Min(1.0, math.Max(0.0, normal.Dot(L.Add(view).Normalize()))), 1/obj.Finish().roughness))
	ambient := light.color.Mult(obj.Color().Scale(obj.Finish().ambient))
	return diffuse.Add(specular).Add(ambient)
}
