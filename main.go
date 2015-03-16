package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
)

var (
	/* For chromebook, save to downloads so file can be viewed from chromeos */
	filename = os.Getenv("HOME") + "/Downloads/trace.png"

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
	pt := Point3D{X: 1, Y: 1, Z: 1}
	pt = pt.Transform(CreateScale(2, 3, 4))
	pt = pt.Transform(CreateTranslate(-5, -6, -7))

	fmt.Println("New Point:", pt)
	return

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("missing .pov file argument")
		return
	}

	povFile, err := os.Open(args[0])
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

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		file.Close()
	}()

	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}
}

func castRay(ray Ray, depth int) (bool, fColor) {
	if depth > MAX_DEPTH {
		return false, bkgndColor
	}

	if count, t1, _, ndx := hitAnything(ray); count > 0 {
		pxlClr := fColor{}
		for i := range lights {
			light := lights[i]
			interPt := ray.PointAt(t1 - 0.01)
			if !isShadowed(interPt, light, ndx) {
				normal := objects[ndx].Normal(interPt)
				reflection := ray.Direction.Sub(normal.Scale(2 * ray.Direction.Dot(normal)))
				pxlClr = pxlClr.Add(calcColor(objects[ndx], light, interPt, eye.location))
				if reflect, color := castRay(Ray{interPt, reflection.Normalize()}, depth+1); reflect {
					rScale := objects[ndx].Finish().reflection
					pxlClr = pxlClr.Add(color.Scale(rScale))
				}
				//				if objects[ndx].Material().Ambient.A < float64(1) {
				//					/*TODO: calc dot product instead of sin */
				//					refract := Ray{Origin: ray.PointAt(t2 + 0.01), Direction: ray.Direction}
				//					_, refractColor := castRay(refract, depth+1)
				//					pxlClr = pxlClr.Scale(objects[ndx].Material().Ambient.A).Add(refractColor.Scale(1 - objects[ndx].Material().Ambient.A))
				//				}
			} else {
				pxlClr = pxlClr.Add(light.color.Mult(objects[ndx].Color().
					Scale(objects[ndx].Finish().ambient)))
			}
		}
		return true, pxlClr
	}
	return false, bkgndColor
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
