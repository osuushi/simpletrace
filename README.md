# simpletrace

**Note:** This is in a prototype stage and I've pushed it for use in my own
projects. I don't recommend depending on it yet. It has inadequate tests, and
needs significant refactoring.

`simpletrace` is a Go package for converting images into polygons. It is similar
to tools like [potrace](http://potrace.sourceforge.net/), but produces simple
polygons, rather than curves, which makes it easier to work with, for example,
in applications where you need to produce polygon meshes.

Tracing is achieved by using [marching
squares](https://en.wikipedia.org/wiki/Marching_squares) to partition the field,
and then simplifying the contours to reduce the line count. Filled regions turn into counterclockwise-wound, simple polygons, and holes are turned into clockwise-wound simple polygons.

simpletrace only supports bitmap tracing, and cannot be extended easily to
handle other colors. It does, however, support converting images to bitmap
through the `IsColorFilledFunc` callback.