package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/fogleman/gg"
	"github.com/osuushi/simpletrace"
)

func main() {
	// Load the image from stdin
	image, err := LoadImageFromStdin()
	if err != nil {
		panic(err)
	}

	polygons := simpletrace.TraceImage(image, simpletrace.DarkColorFilledFunc)
	fmt.Println(len(polygons), "polygons found")
	fmt.Println("Polygon sizes:")
	for _, polygon := range polygons {
		fmt.Println(len(polygon), "points")
	}

	// Draw the polygons into an image
	cx := gg.NewContext(image.Bounds().Dx()*4, image.Bounds().Dy()*4)
	fmt.Println("Created context")
	cx.SetRGB255(0, 0, 0)
	cx.Clear()
	fmt.Println("Drawing polygons")
	for _, polygon := range polygons {
		isHole := simpletrace.SignedAreaOfPolygon(polygon) < 0
		if isHole {
			cx.SetRGB255(255, 0, 0)
		} else {
			cx.SetRGB255(0, 0, 255)
		}
		for i := 0; i < len(polygon); i++ {
			nextI := (i + 1) % len(polygon)
			// Draw scaled up by 4 to get a better view of the lines
			cx.DrawLine(polygon[i].X*4, polygon[i].Y*4, polygon[nextI].X*4, polygon[nextI].Y*4)
		}
		cx.Stroke()

		// Draw green points at the end points of the lines
		for _, p := range polygon {
			cx.SetRGB255(0, 255, 0)
			cx.DrawCircle(p.X*4, p.Y*4, 1)
		}
		cx.Fill()
	}
	fmt.Println("saving image to demo.png")
	cx.SavePNG("demo.png")

	// Reproduce the original image
	cx = gg.NewContext(image.Bounds().Dx(), image.Bounds().Dy())
	cx.SetRGB255(255, 255, 255)
	cx.Clear()
	cx.SetFillRuleWinding()
	for _, polygon := range polygons {
		cx.MoveTo(polygon[0].X, polygon[0].Y)
		for i := 1; i < len(polygon); i++ {
			cx.LineTo(polygon[i].X, polygon[i].Y)
		}
		cx.ClosePath()
	}
	cx.SetRGB255(0, 0, 0)
	cx.Fill()
	cx.SavePNG("demo-reproduced.png")
}

func LoadImageFromStdin() (image.Image, error) {
	img, _, err := image.Decode(os.Stdin)
	return img, err
}
