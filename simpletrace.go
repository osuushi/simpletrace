package simpletrace

import (
	"image"
)

func TraceImage(img image.Image, isColorFilledFunc IsColorFilledFunc) [][]Point {
	// Make the square map
	squaremap := getSquaresForImage(img, isColorFilledFunc)
	// Get the polygons
	polygons := squaremap.convertSquaresToPolygons()
	return polygons
}
