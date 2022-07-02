package simpletrace

import (
	"image"

	"github.com/kr/pretty"
)

func TraceImage(img image.Image) [][]Point {
	// Make the square map
	squaremap := getSquaresForImage(img, DarkColorFilledFunc)
	pretty.Println("Squares", squaremap.Inspect())
	// Get the polygons
	polygons := squaremap.convertSquaresToPolygons()
	return polygons
}
