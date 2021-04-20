package gfx

import "fmt"

type BoundingBox struct {
	X, Y, Z int
	W, H, D int
}

func (bb *BoundingBox) Set(x, y, z, w, h, d int) {
	bb.SetPos(x, y, z)
	bb.W = w
	bb.H = h
	bb.D = d
}

func (bb *BoundingBox) SetPos(x, y, z int) {
	bb.X = x
	bb.Y = y
	bb.Z = z
}

func (bb *BoundingBox) isInside(x, y, z int) bool {
	return bb.X <= x && x < bb.X+bb.W &&
		bb.Y <= y && y < bb.Y+bb.H &&
		bb.Z <= z && z < bb.Z+bb.D
}

func sideOverlap(ax1, ax2, bx1, bx2 int) bool {
	return ax1 < bx2 && ax2 > bx1
}

func (a *BoundingBox) intersect(b *BoundingBox) bool {
	return sideOverlap(a.X, a.X+a.W, b.X, b.X+b.W) &&
		sideOverlap(a.Y, a.Y+a.H, b.Y, b.Y+b.H) &&
		sideOverlap(a.Z, a.Z+a.D, b.Z, b.Z+b.D)
}

func (bp *BoundingBox) describe() string {
	return fmt.Sprintf("%d,%d,%d-%d,%d,%d", bp.X, bp.Y, bp.Z, bp.X+bp.W, bp.Y+bp.H, bp.Z+bp.D)
}
