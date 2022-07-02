package simpletrace

import (
	"math"
	"strings"
)

type SquareMap map[IPoint]*Square

type RotationMatrix [4][4]float64

func (s SquareMap) convertSquaresToPolygons() [][]Point {
	var polygons [][]Point

	// Grab random squares and then consume their neighbors
	for len(s) > 0 {
		startingSquare := s.GetSquare()
		polygon := s.tracePolygonFromSquare(startingSquare)
		polygons = append(polygons, polygon)
	}
	return polygons
}

// Get any square from the map.
func (s SquareMap) GetSquare() *Square {
	for p := range s {
		return s[p]
	}
	panic("No squares left")
}

// Check if the square has lost all its edges. This is to deal with saddle
// points, which must be converted to non-saddle points once the first path
// through them is handled.
func (s SquareMap) garbageCollect(square *Square) {
	if square.Corners == CornerStateNone || square.Corners == CornerStateAll {
		delete(s, square.Point)
	}
}

func (s SquareMap) tracePolygonFromSquare(startingSquare *Square) []Point {
	var polygon []Point
	var currentDirection Direction
	var startPointDirection Direction
	var lastDirection Direction
	lastSquare := startingSquare

	// Find some direction that redirects to a neighbor
	var foundDirection bool
	for direction := Direction(0); direction < DirectionInvalid; direction++ {
		currentDirection = lastSquare.DirectionFor(direction)
		if currentDirection != DirectionInvalid {
			foundDirection = true
			startPointDirection = direction
			break
		}
	}

	if !foundDirection {
		panic("No valid direction found for square " + startingSquare.Inspect())
	}

	// The starting point will be the midpoint of the two exit sides for the direction we chose
	// above.

	a, b := lastSquare.CornerPointsInDirection(startPointDirection)
	segmentStart := Point{(a.X + b.X) / 2, (a.Y + b.Y) / 2}

	// In order to determine when we need to start a new line segment, we will
	// track the bounds of the angle that the current line segment is constrained
	// to. The rule is that the line segment cannot "escape" the two corners it
	// passes through as it exits a square. Otherwise, we can simplify the path
	// through as many squares as we want to a single segment.
	//
	// However, dealing with angles here would be messy and require unwinding, so
	// instead, we transform the coordinate system to put the segment's initial
	// direction on the x-axis. We can then work entirely in unit vectors pointing
	// from the start point to each corner we encounter, and simply compare the y
	// values.

	// State set up by setUpNextSegment, and managed during the loop
	var rotationMatrix RotationMatrix
	var inverseRotationMatrix RotationMatrix
	var minY float64
	var maxY float64

	var setUpNextSegment = func(square *Square, direction Direction) {
		// The segment's initial direction is the vector from the start point to the
		// midpoint of the two exit corners.

		// Find the midpoint of the two exit corners. That's what we'll rotate to the x-axis.
		a, b := square.CornerPointsInDirection(direction)
		a, b = squeezeCorners(a, b)
		midpoint := Point{(a.X + b.X) / 2, (a.Y + b.Y) / 2}
		baseline := segmentStart.UnitVectorTo(midpoint)

		// Find the rotation matrix that will transform the segment's initial direction to the x-axis.
		rotationMatrix = baselineToXAxis(baseline)
		inverseRotationMatrix = rotationMatrix.invert()

		// Get the two corner unit vectors
		a = segmentStart.UnitVectorTo(a)
		b = segmentStart.UnitVectorTo(b)

		a = rotationMatrix.multiply(a)
		b = rotationMatrix.multiply(b)

		// Get the min and max y values
		minY = math.Min(a.Y, b.Y)
		maxY = math.Max(a.Y, b.Y)
	}

	// When we terminate a segment, we have to find the endpoint. We do this by
	// splitting the constraint wedge, and finding the intersection with the
	// outgoing edge of the last square
	var findEndOfCurrentSegment = func() Point {
		// Find where the middle of the constraints hits the previous edge
		midY := (minY + maxY) / 2
		segmentAngle := math.Asin(midY)
		segmentVector := Point{math.Cos(segmentAngle), math.Sin(segmentAngle)}

		// Transform back to the original coordinate system
		segmentVector = inverseRotationMatrix.multiply(segmentVector)

		// We need just one incoming corner, because we know its orientation from the direction
		lastExitingCornerA, lastExitingCornerB := lastSquare.CornerPointsInDirection(lastDirection)
		var endpoint Point

		// Find where the line at the segment angle from the segment start intersects the incoming edge
		if lastDirection.IsVertical() {
			// If the direction is vertical, the edge is horizontal
			// Check if this is the same row. If so, we can't solve for x, and just take the midpoint of the corners
			if math.Abs(segmentVector.Y) < 1e-6 {
				endpoint = Point{(lastExitingCornerA.X + lastExitingCornerB.X) / 2, segmentStart.Y}
			} else {
				// Solve for x
				endpoint.Y = lastExitingCornerA.Y
				segmentLength := (endpoint.Y - segmentStart.Y) / segmentVector.Y
				endpoint.X = segmentStart.X + segmentVector.X*segmentLength
			}
		} else {
			// If the direction is horizontal, the edge is vertical
			// Check if this is the same column. If so, we can't solve for y, and just take the midpoint of the corners
			if math.Abs(segmentVector.X) < 1e-6 {
				endpoint = Point{segmentStart.X, (lastExitingCornerA.Y + lastExitingCornerB.Y) / 2}
			} else {
				endpoint.X = lastExitingCornerA.X
				segmentLength := (endpoint.X - segmentStart.X) / segmentVector.X

				endpoint.Y = segmentStart.Y + segmentVector.Y*segmentLength
			}
		}
		return endpoint
	}

	// Save the state of the top left corner of the square. By checking if that
	// corner is included in the polygon, we will learn if this is a hole or not.
	startingSquareContent := *startingSquare
	topLeftMostSquareContent := *startingSquare
	var leftMostNonSaddle *Square // must always be a copy because of cleanup
	if !startingSquare.Corners.IsSaddle() {
		leftMostNonSaddle = &startingSquareContent
	}

	setUpNextSegment(lastSquare, currentDirection)
	polygonStart := segmentStart

	for {
		lastDirection = currentDirection

		// Get the neighbor for the current square
		currentIPoint := lastSquare.Point.ApplyDirection(currentDirection)

		// Stop when we come back to the start
		if currentIPoint == startingSquare.Point {
			break
		}

		currentSquare := s[currentIPoint]

		// Check topleftmost square values
		if currentIPoint.X < topLeftMostSquareContent.Point.X {
			topLeftMostSquareContent = *currentSquare
		} else if currentIPoint.X == topLeftMostSquareContent.Point.X && currentIPoint.Y < topLeftMostSquareContent.Point.Y {
			topLeftMostSquareContent = *currentSquare
		}

		// Check leftmost non-saddle square values
		if currentIPoint.X < leftMostNonSaddle.Point.X && !currentSquare.Corners.IsSaddle() {
			square := *currentSquare
			leftMostNonSaddle = &square
		}

		// Get the new direction
		currentDirection = currentSquare.DirectionFor(currentDirection)

		// Get the corner points for exiting the new square
		exitA, exitB := currentSquare.CornerPointsInDirection(currentDirection)

		proposedExit := Point{(exitA.X + exitB.X) / 2, (exitA.Y + exitB.Y) / 2}
		segmentToExit := segmentStart.UnitVectorTo(proposedExit)
		segmentToExit = rotationMatrix.multiply(segmentToExit)

		// Check if the exit is outside the constraining wedge. If so, we have to
		// end the segment at the entrance, because the exit would not be a valid end to the current
		// segment.
		if segmentToExit.Y < minY || segmentToExit.Y > maxY {
			// End the current segment at the entrance to this square
			entranceA, entranceB := currentSquare.CornerPointsInDirection(lastDirection.Reverse())
			entrance := Point{(entranceA.X + entranceB.X) / 2, (entranceA.Y + entranceB.Y) / 2}

			polygon = append(polygon, entrance)

			// Start the next segment
			segmentStart = entrance
			setUpNextSegment(currentSquare, currentDirection)
		} else {
			// Update the constraining wedge
			exitA, exitB = squeezeCorners(exitA, exitB)
			exitA = segmentStart.UnitVectorTo(exitA)
			exitB = segmentStart.UnitVectorTo(exitB)
			exitA = rotationMatrix.multiply(exitA)
			exitB = rotationMatrix.multiply(exitB)

			minY = math.Max(math.Min(exitA.Y, exitB.Y), minY)
			maxY = math.Min(math.Max(exitA.Y, exitB.Y), maxY)
		}

		// Clean up the last square
		lastSquare.RemovePathForOutgoingDirection(lastDirection)
		s.garbageCollect(lastSquare)
		lastSquare = currentSquare
	}

	// Clean up the starting square
	lastSquare.RemovePathForOutgoingDirection(lastDirection)
	s.garbageCollect(lastSquare)

	// Check if the starting point for the polygon is an acceptable end point for
	// the last segment. If not, we need to make one last tiny segment to join up the polygon
	backToStartVector := segmentStart.UnitVectorTo(polygonStart)
	if backToStartVector.Y > maxY || backToStartVector.Y < minY { // We need to add a connecting segment
		endpoint := findEndOfCurrentSegment()
		polygon = append(polygon, endpoint)
	}

	// Determine if we need to reverse the polygon so that counterclockwise = filled

	var polygonIsFilled bool
	if leftMostNonSaddle == nil {
		// If there are no non-saddle points in the polygon, then this is a single
		// pixel case, so the top left square's lower right corner tells us if the polygon is filled
		polygonIsFilled = topLeftMostSquareContent.Corners&CornerStateBottomRight != 0
	} else {
		// Otherwise, the left most non-saddle point's top left corner tells us if
		// the polygon is filled, since it is guaranteed to be outside the polygon
		polygonIsFilled = leftMostNonSaddle.Corners&CornerStateTopLeft == 0
	}

	signedArea := SignedAreaOfPolygon(polygon)
	if polygonIsFilled != (signedArea > 0) {
		reversePolygon(polygon)
	}

	return polygon
}

// Create a rotation matrix that will rotate baseline to {X, 0} for some positive X
func baselineToXAxis(baseline Point) RotationMatrix {
	angle := math.Atan2(baseline.Y, baseline.X)
	return RotationMatrix{
		{math.Cos(angle), math.Sin(angle)},
		{-math.Sin(angle), math.Cos(angle)},
	}
}

func (m RotationMatrix) multiply(p Point) Point {
	return Point{
		m[0][0]*p.X + m[0][1]*p.Y,
		m[1][0]*p.X + m[1][1]*p.Y,
	}
}

// Invert the rotation matrix to get the rotation in the other direction
func (m RotationMatrix) invert() RotationMatrix {
	var a, b, c, d = m[0][0], m[0][1], m[1][0], m[1][1]
	return RotationMatrix{
		{d, -b},
		{-c, a},
	}
}

// Slightly nudge corners toward each other to ensure that polygons will never
// touch each other. This is used when computing the constarints of a segment.
func squeezeCorners(a, b Point) (Point, Point) {
	const squeezeFactor = 1 / 8.0
	a = Point{a.X*(1-squeezeFactor) + b.X*squeezeFactor, a.Y*(1-squeezeFactor) + b.Y*squeezeFactor}
	b = Point{b.X*(1-squeezeFactor) + a.X*squeezeFactor, b.Y*(1-squeezeFactor) + a.Y*squeezeFactor}
	return a, b
}

func SignedAreaOfPolygon(polygon []Point) float64 {
	area := 0.0
	n := len(polygon)
	for i := 0; i < n; i++ {
		nextI := (i + 1) % n
		area += polygon[i].X*polygon[nextI].Y - polygon[nextI].X*polygon[i].Y
	}
	return area / 2
}

func reversePolygon(polygon []Point) []Point {
	for i, j := 0, len(polygon)-1; i < j; i, j = i+1, j-1 {
		polygon[i], polygon[j] = polygon[j], polygon[i]
	}
	return polygon
}

func (sm SquareMap) Inspect() string {
	var sb strings.Builder
	for _, square := range sm {
		sb.WriteString(square.Inspect())
	}
	return sb.String()
}
