package simpletrace

import (
	"image"
)

func TraceImage(img image.Image) [][]Point {
	// Make the square map
	squaremap := getSquaresForImage(img, DarkColorFilledFunc)
	// Get the polygons
	polygons := squaremap.convertSquaresToPolygons()
	return polygons
}
