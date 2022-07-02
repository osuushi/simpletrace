package simpletrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaselineToXAxis(t *testing.T) {
	cases := []Point{
		{.3, .5},
		{-.3, .8},
		{7, -5},
		{0, -5},
		{0, 2},
		{3, 0},
		{-10, 0},
	}

	for _, c := range cases {
		actual := baselineToXAxis(c).multiply(c)
		assert.InDeltaf(t, 0, actual.Y, 0.00001, "expected (%.2f, %.2f) to rotate to the x axis, but got (%.2f, %.2f)", c.X, c.Y, actual.X, actual.Y)
		assert.Greater(t, actual.X, 0.0, "expected %v to be on the positive x axis, but got (%.2f, %.2f)", c, actual.X, actual.Y)

		// Invert the matrix to get the rotation in the other direction
		inverse := baselineToXAxis(c).invert()
		actual = inverse.multiply(actual)
		assert.InDeltaf(t, c.X, actual.X, 0.00001, "expected %v to be rotated back to %v, but got (%.2f, %.2f)", c, c, actual.X, actual.Y)
		assert.InDeltaf(t, c.Y, actual.Y, 0.00001, "expected %v to be rotated back to %v, but got (%.2f, %.2f)", c, c, actual.X, actual.Y)
	}
}
