package simpletrace

import "math"

type Point struct {
	X float64
	Y float64
}

type IPoint struct {
	X int
	Y int
}

type Direction uint8

const (
	DirectionUp = Direction(iota)
	DirectionRight
	DirectionDown
	DirectionLeft
	DirectionInvalid
)

func PointFromInts(x, y int) Point {
	return Point{float64(x), float64(y)}
}

func (p IPoint) ApplyDirection(dir Direction) IPoint {
	switch dir {
	case DirectionUp:
		return IPoint{p.X, p.Y - 1}
	case DirectionRight:
		return IPoint{p.X + 1, p.Y}
	case DirectionDown:
		return IPoint{p.X, p.Y + 1}
	case DirectionLeft:
		return IPoint{p.X - 1, p.Y}
	}
	panic("Invalid direction")
}

func (dir Direction) Reverse() Direction {
	return (dir + 2) % 4
}

func (dir Direction) IsVertical() bool {
	return dir == DirectionUp || dir == DirectionDown
}

func (dir Direction) String() string {
	switch dir {
	case DirectionUp:
		return "Up"
	case DirectionRight:
		return "Right"
	case DirectionDown:
		return "Down"
	case DirectionLeft:
		return "Left"
	case DirectionInvalid:
		return "INVALID"
	}
	panic("Invalid direction")
}

func (p Point) UnitVectorTo(other Point) Point {
	result := Point{
		X: other.X - p.X,
		Y: other.Y - p.Y,
	}
	len := math.Sqrt(result.X*result.X + result.Y*result.Y)
	result.X /= len
	result.Y /= len
	return result
}
