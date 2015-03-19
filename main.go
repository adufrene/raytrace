package main

import (
	"fmt"
	"image"
	//	"image/jpeg"
	"image/png"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	/* For chromebook, save to downloads so file can be viewed from chromeos */
	fileDir = os.Getenv("HOME") + "/Downloads/"
	//	ext     = ".jpg"
	ext = ".png"

	imgWidth  = 800
	imgHeight = 600
	//	imgWidth  = 20
	//	imgHeight = 14

	red   = fColor{R: 1.0, G: 0.0, B: 0.0, A: 1.0}
	green = fColor{R: 0.0, G: 0.0, B: 0.0, A: 1.0}
	blue  = fColor{R: 0.0, G: 0.0, B: 1.0, A: 1.0}
	white = fColor{R: 1.0, G: 1.0, B: 1.0, A: 1.0}
	black = fColor{R: 0.0, G: 0.0, B: 0.0, A: 1.0}

	bkgndColor = black // fColor{R: 0.2706, G: 0.3137, B: 0.3294, A: 1.0}

	MAX_DEPTH  = 7
	numThreads int
)

type goArgs struct {
	ray  Ray
	x, y int
}

func main() {
	povFile := processCmd()
	if povFile == nil {
		return
	}
	defer povFile.Close()
	argsChan := make(chan goArgs, 4096)
	img := image.NewRGBA(image.Rectangle{image.ZP, image.Point{imgWidth, imgHeight}})
	wg := sync.WaitGroup{}
	setupThreads(argsChan, &wg, img)

	xTrans := eye.right.Scale(2 / float64(imgWidth))
	yTrans := eye.up.Scale(2 / float64(imgHeight))

	xStart := eye.location.Translate(eye.right.Scale(-1))
	yStart := eye.location.Translate(eye.up.Scale(-1))
	imgPlane := eye.lookAt.Sub(eye.location).Normalize().Scale(2)

	currX := xStart.Translate(xTrans.Scale(0.5))
	for x := 0; x < imgWidth; x++ {
		currY := yStart.Translate(yTrans.Scale(0.5))
		for y := imgHeight - 1; y >= 0; y-- {
			view := eye.location.Translate(currX.Sub(eye.location)).
				Translate(currY.Sub(eye.location)).Translate(imgPlane)
			argsChan <- goArgs{CreateRay(eye.location, view), x, y}
			currY = currY.Translate(yTrans)
		}
		currX = currX.Translate(xTrans)
	}
	close(argsChan)
	wg.Wait()

	writeFile(img)
}

func processCmd() *os.File {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage:", os.Args[0], "<path-to-pov-file>")
		return nil
	}

	filename := args[0]

	povFile, err := os.Open(filename)
	defer povFile.Close()
	if err == nil {
		err = parsePOV(povFile)
	}
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return povFile
}

func setupThreads(channel chan goArgs, wg *sync.WaitGroup, img *image.RGBA) {
	maxProcsString := os.Getenv("GOMAXPROCS")
	if maxProcsString == "" {
		numThreads = runtime.NumCPU()
	} else {
		numThreads64, err := strconv.ParseInt(maxProcsString, 10, 32)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		numThreads = int(numThreads64)
	}
	runtime.GOMAXPROCS(int(numThreads))
	fmt.Println("Using", numThreads, "thread(s)")

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			for arg := range channel {
				_, color := castRay(arg.ray, MAX_DEPTH, -1)
				img.Set(arg.x, arg.y, color)
			}
			wg.Done()
		}()
	}
}

func writeFile(img *image.RGBA) {
	splitString := strings.Split(os.Args[1], "/")
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

	//	err = jpeg.Encode(file, img, nil)
	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}
}

func castRay(ray Ray, depth, currObj int) (bool, fColor) {
	depth--
	if depth < 0 {
		return false, bkgndColor
	}

	if hit, t, ndx := hitAnything(ray, currObj); hit {
		obj := objects[ndx]
		pxlClr := fColor{}
		interPt := ray.PointAt(t)
		for i := range lights {
			light := lights[i]
			if !isShadowed(interPt, light, ndx) {
				pxlClr = pxlClr.Add(calcColor(obj, light, interPt, eye.location))
			} else {
				pxlClr = pxlClr.Add(light.color.Mult(obj.Color().
					Scale(obj.Finish().ambient)))
			}
		}
		normal := obj.Normal(interPt)
		if obj.Finish().reflection > 0 {
			reflection := ray.Direction.Sub(normal.Scale(2 * ray.Direction.Dot(normal)))
			if reflect, color := castRay(Ray{interPt, reflection.Normalize()}, depth, ndx); reflect {
				pxlClr = pxlClr.Add(color.Scale(obj.Finish().reflection))
			}
		}
		if obj.Finish().refraction > 0 {
			// Assuming non object material is air w/ ior=1
			var internal bool
			var refractRay Ray
			//					internal, refractRay = calcRefractRay(ray, obj, origPt, 1, 1)
			if ray.Direction.Dot(normal) > 0 { // We are exiting the object
				internal, refractRay = calcRefractRay(ray, obj, interPt, obj.Finish().ior, 1)
			} else { // We are entering the object
				internal, refractRay = calcRefractRay(ray, obj, interPt, 1, obj.Finish().ior)
			}
			if !internal {
				if refract, color := castRay(refractRay, depth, ndx); refract {
					pxlClr = pxlClr.Add(color.Scale(obj.Finish().refraction))
				}
			}
		}
		if pxlClr.A < 1 {
			_, nextClr := castRay(ray, depth, ndx)
			pxlClr = pxlClr.Scale(pxlClr.A).Add(nextClr.Scale(1 - pxlClr.A))
		}
		return true, pxlClr
	}
	return false, bkgndColor
}

func calcRefractRay(initialRay Ray, obj castable, origPt Point3D,
	n1, n2 float64) (internalReflection bool, refractRay Ray) {
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
			if hitObj, _ := objects[ndx].Hit(r); hitObj {
				return true
			}
		}
	}
	return false
}

func hitAnything(r Ray, objNdx int) (hit bool, t float64, hitNdx int) {
	t = math.MaxFloat64
	for ndx := range objects {
		if ndx != objNdx {
			if hitObj, hitTime := objects[ndx].Hit(r); hitObj && hitTime < t {
				hit, t, hitNdx = true, hitTime, ndx
			}
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
