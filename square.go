package simpletrace

import (
	"fmt"

	"github.com/lithammer/dedent"
)

// A single square for the marching squares algorithm.
type Square struct {
	Point   IPoint
	Corners CornerStates
}

// Whether each corner of the square is set in the original bitmap
type CornerStates uint8

const (
	CornerStateTopLeft = CornerStates(1 << iota)
	CornerStateTopRight
	CornerStateBottomLeft
	CornerStateBottomRight
)
const CornerStateNone = CornerStates(0)
const CornerStateAll = CornerStateTopLeft | CornerStateTopRight | CornerStateBottomLeft | CornerStateBottomRight

var SaddleCornerStates = [2]CornerStates{
	CornerStateTopLeft | CornerStateBottomRight,
	CornerStateTopRight | CornerStateBottomLeft,
}

func (s Square) DirectionForNeighbor(neighbor Square) Direction {
	if neighbor.Point.X == s.Point.X {
		if neighbor.Point.Y < s.Point.Y {
			return DirectionUp
		} else {
			return DirectionDown
		}
	} else if neighbor.Point.Y == s.Point.Y {
		if neighbor.Point.X < s.Point.X {
			return DirectionLeft
		} else {
			return DirectionRight
		}
	} else {
		panic("Invalid neighbor")
	}
}

func CornerStateForOffset(x, y int) CornerStates {
	return 1 << (y*2 + x)
}

// This is where the actual marching squares rules take place, encoding the
// change in direction of a path through the square. For example, a square like
// this:
//
//  X----\-O
//  |     \|
//  |      \
//  |      |
//  X------X
//
// Will return DirectionRight if passed DirectionDown, and DirectionUp if passed
// DirectionLeft, because when traveling downward, you are redirected to the
// right, and so on.

func (s Square) DirectionFor(from Direction) Direction {
	return Redirections[s.Corners][from]
}

func (s Square) CornerPointsInDirection(dir Direction) (Point, Point) {
	switch dir {
	case DirectionUp:
		return PointFromInts(s.Point.X, s.Point.Y), PointFromInts(s.Point.X+1, s.Point.Y)
	case DirectionRight:
		return PointFromInts(s.Point.X+1, s.Point.Y), PointFromInts(s.Point.X+1, s.Point.Y+1)
	case DirectionDown:
		return PointFromInts(s.Point.X+1, s.Point.Y+1), PointFromInts(s.Point.X, s.Point.Y+1)
	case DirectionLeft:
		return PointFromInts(s.Point.X, s.Point.Y+1), PointFromInts(s.Point.X, s.Point.Y)
	default:
		panic("Invalid direction")
	}
}

// For an outgoing direction, remove the path by updating the corner states.
// This allows saddle points to be handled correctly, since they have two paths.
// If a square loses all its paths, it gets garbage collected
func (s *Square) RemovePathForOutgoingDirection(outgoingDirection Direction) {
	if !s.Corners.IsSaddle() {
		// If there's no saddle, we don't care about the specifics. There's only one
		// path, so we delete it by blanking the corners.
		s.Corners = 0
		return
	}

	// For saddle points, if the path goes through the bottom edge, we can
	// eliminate it by blanking both bottom corners. Otherwise, we blank both top
	// corners.
	if outgoingDirection == DirectionDown || s.DirectionFor(outgoingDirection.Reverse()) == DirectionDown {
		s.Corners &= (^CornerStateBottomLeft & ^CornerStateBottomRight)
	} else {
		s.Corners &= (^CornerStateTopLeft & ^CornerStateTopRight)
	}
}

func (s *Square) Inspect() string {
	formatArgs := []interface{}{s.Point.X, s.Point.Y}
	for _, corner := range []CornerStates{CornerStateTopLeft, CornerStateTopRight, CornerStateBottomLeft, CornerStateBottomRight} {
		if s.Corners&corner == corner {
			formatArgs = append(formatArgs, "X")
		} else {
			formatArgs = append(formatArgs, " ")
		}
	}

	return fmt.Sprintf(dedent.Dedent(`
		Square:
			Point: (%d, %d)
			Corners:
				|%s %s|
				|%s %s|
	`), formatArgs...)
}

// Lookup table for the marching squares
var Redirections [16][4]Direction

func init() {
	// Start all the redirections with invalid
	for i := range Redirections {
		for j := range Redirections[i] {
			Redirections[i][j] = DirectionInvalid
		}
	}

	// Single corner cases (diagonals)
	setRedirection(CornerStateTopLeft, DirectionDown, DirectionLeft)
	setRedirection(CornerStateTopRight, DirectionDown, DirectionRight)
	setRedirection(CornerStateBottomLeft, DirectionUp, DirectionLeft)
	setRedirection(CornerStateBottomRight, DirectionUp, DirectionRight)

	// Two corner same side cases (horizontal/vertical passthrough)
	setRedirection(CornerStateTopLeft|CornerStateTopRight, DirectionLeft, DirectionLeft)       // tops
	setRedirection(CornerStateBottomLeft|CornerStateBottomRight, DirectionLeft, DirectionLeft) // bottoms
	setRedirection(CornerStateTopLeft|CornerStateBottomLeft, DirectionUp, DirectionUp)         // lefts
	setRedirection(CornerStateTopRight|CornerStateBottomRight, DirectionUp, DirectionUp)       // rights

	// Two corner different side cases (ambiguous case, but we pick the
	// "disconnected" version, where the lines separate the two filled corners)
	//
	// Note that this is two cases, because each case has two paths.
	setRedirection(CornerStateTopLeft|CornerStateBottomRight, DirectionRight, DirectionUp)
	setRedirection(CornerStateTopLeft|CornerStateBottomRight, DirectionLeft, DirectionDown)
	setRedirection(CornerStateTopRight|CornerStateBottomLeft, DirectionLeft, DirectionUp)
	setRedirection(CornerStateTopRight|CornerStateBottomLeft, DirectionRight, DirectionDown)

	// Three corner cases
	setRedirection(CornerStateTopLeft|CornerStateTopRight|CornerStateBottomLeft, DirectionUp, DirectionRight)
	setRedirection(CornerStateTopLeft|CornerStateTopRight|CornerStateBottomRight, DirectionUp, DirectionLeft)
	setRedirection(CornerStateTopLeft|CornerStateBottomLeft|CornerStateBottomRight, DirectionLeft, DirectionUp)
	setRedirection(CornerStateTopRight|CornerStateBottomLeft|CornerStateBottomRight, DirectionRight, DirectionUp)

	// Remaining two cases are all filled, or all empty, which means there's no path through them
}

// Helper to make the lookup table setup easier by setting both directions at once
func setRedirection(states CornerStates, from Direction, to Direction) {
	Redirections[states][from] = to
	Redirections[states][to.Reverse()] = from.Reverse()
}

func (cs CornerStates) IsSaddle() bool {
	return cs == CornerStateTopLeft|CornerStateBottomRight || cs == CornerStateTopRight|CornerStateBottomLeft
}
