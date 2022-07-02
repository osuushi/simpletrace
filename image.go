package simpletrace

import (
	"image"
	"image/color"
)

// Callback for determining if a pixel is filled
type IsColorFilledFunc func(color.Color) bool

// Convert to bitmap by 50% alpha threshold
var OpacityColorFilledFunc IsColorFilledFunc = func(c color.Color) bool {
	_, _, _, alpha := c.RGBA()
	return alpha > 0xffff/2
}

// Convert to bitmap by 50% lightness threshold, where black is filled
var DarkColorFilledFunc IsColorFilledFunc = func(c color.Color) bool {
	if !OpacityColorFilledFunc(c) {
		return false
	}
	y := yValueFromColor(c)
	return y < 0x80
}

// Convert to a bitmap by 50% lightness threshold, where white is filled
var LightColorFilledFunc IsColorFilledFunc = func(c color.Color) bool {
	if !OpacityColorFilledFunc(c) {
		return false
	}
	y := yValueFromColor(c)
	return y > 0x80
}

func yValueFromColor(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	// Convert to uint8
	r = r >> 8
	g = g >> 8
	b = b >> 8
	// Convert to ycbcr
	y, _, _ := color.RGBToYCbCr(uint8(r), uint8(g), uint8(b))
	return y
}

// Convert an image into a set of squares for
func getSquaresForImage(img image.Image, isColorFilled IsColorFilledFunc) SquareMap {
	squares := make(SquareMap)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Convert the alpha channel of the 2x2 square here to the corner states
			// of the square.
			corners := CornerStates(0)
			for offsetY := 0; offsetY < 2; offsetY++ {
				for offsetX := 0; offsetX < 2; offsetX++ {
					if isColorFilled(img.At(x+offsetX, y+offsetY)) {
						corners |= CornerStateForOffset(offsetX, offsetY)
					}
				}
			}

			// Squares that contain no edges are omitted from the map
			if corners == CornerStateNone || corners == CornerStateAll {
				continue
			}
			square := Square{
				Point:   IPoint{x, y},
				Corners: corners,
			}

			squares[square.Point] = &square
		}
	}
	return squares
}
