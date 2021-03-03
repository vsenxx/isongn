package shapes

type Direction string

var DIR_W = Direction("w")

var DIR_SW = Direction("sw")
var DIR_S = Direction("s")
var DIR_SE = Direction("se")
var DIR_E = Direction("e")
var DIR_NE = Direction("ne")
var DIR_N = Direction("n")
var DIR_NW = Direction("nw")
var DIR_NONE = Direction("none")

var Directions = []*Direction{
	&DIR_W,
	&DIR_SW,
	&DIR_S,
	&DIR_SE,
	&DIR_E,
	&DIR_NE,
	&DIR_N,
	&DIR_NW,
	&DIR_NONE,
}

func GetDir(oldX, oldY, newX, newY int) *Direction {
	if oldX < newX && oldY == newY {
		return &DIR_W
	}
	if oldX > newX && oldY == newY {
		return &DIR_E
	}
	if oldX == newX && oldY < newY {
		return &DIR_S
	}
	if oldX == newX && oldY > newY {
		return &DIR_N
	}
	if oldX < newX && oldY < newY {
		return &DIR_SW
	}
	if oldX > newX && oldY > newY {
		return &DIR_NE
	}
	if oldX > newX && oldY < newY {
		return &DIR_SE
	}
	if oldX < newX && oldY > newY {
		return &DIR_NW
	}
	return &DIR_NONE
}
