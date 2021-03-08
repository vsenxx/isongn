# the global player state
player := {
    "x": 976,
    "y": 1028,
    "z": 1,
    "scrollOffsetX": 0,
    "scrollOffsetY": 0,
    "dir": DirN,
    "underRoof": false,
};

# the player's shape size
const PLAYER_X = 2;
const PLAYER_Y = 2;
const PLAYER_Z = 4;

const PLAYER_SPEED = 0.085;

const PLAYER_SHAPE = "man";

# called on every frame
def events(delta) {
    animationType := "stand";
    dx := 0;
    dy := 0;
    if(isDown(KeyA) || isDown(KeyLeft)) {
        dx := 1;
    }
    if(isDown(KeyD) || isDown(KeyRight)) {
        dx := -1;
    }
    if(isDown(KeyW) || isDown(KeyUp)) {
        dy := -1;
    }
    if(isDown(KeyS) || isDown(KeyDown)) {
        dy := 1;
    }

    if(dx != 0 || dy != 0) {
        animationType := "move";
        playerMove(dx, dy, delta);
    }

    if(isPressed(KeySpace)) {
        if(findShapeNearby("door.wood.y", (x,y,z) => replaceShape(x, y, z, "door.wood.x")) = false) {
            if(findShapeNearby("door.wood.x", (x,y,z) => replaceShape(x, y, z, "door.wood.y")) = false) {
                # do something else
            }
        }
    }

    setAnimation(player.x, player.y, player.z, animationType, player.dir);
}

def playerMove(dx, dy, delta) {
    moved := playerMoveDir(dx, dy, delta);
    if(moved = false && dx != 0 && dy != 0) {
        moved := playerMoveDir(0, dy, delta);
    }
    if(moved = false && dx != 0 && dy != 0) {
        moved := playerMoveDir(dx, 0, delta);
    }
    if(moved) {
        # if under a roof, hide the roof
        player.underRoof := isUnderRoof();
        if(player.underRoof) {
            setMaxZ(getRoofZ());
        } else {
            setMaxZ(24);
        }
    }
    return moved;
}

def playerMoveDir(dx, dy, delta) {
    oldX := player.x;
    oldY := player.y;
    oldZ := player.z;
    player.dir := getDir(dx, dy);
    moved := true;

    # adjust speed for diagonal movement... maybe this should be computed from iso angles?
    speed := PLAYER_SPEED;
    if(dx != 0 && dy != 0) {
        speed := PLAYER_SPEED * 1.5;
    }
    newXf := player.x + player.scrollOffsetX + (dx * delta / speed);
    newYf := player.y + player.scrollOffsetY + (dy * delta / speed);
    newX := int(newXf + 0.5);
    newY := int(newYf + 0.5);

    if(newX != oldX || newY != oldY) {
        eraseShape(oldX, oldY, oldZ);
        newZ := findTopFit(newX, newY, PLAYER_SHAPE);
        if(newZ <= player.z + 1 && inspectUnder(newX, newY, newZ)) {
            moveViewTo(newX, newY);

            player.x := newX;
            player.y := newY;
            player.z := newZ;
        } else {
            # player is blocked
            moved := false;
        }
        setShape(player.x, player.y, player.z, PLAYER_SHAPE);
    }

    if(moved) {
        # pixel scrolling
        player.scrollOffsetX := newXf - newX;
        player.scrollOffsetY := newYf - newY;
        setViewScroll(player.scrollOffsetX, player.scrollOffsetY);
        setOffset(player.x, player.y, player.z, player.scrollOffsetX, player.scrollOffsetY);
    }
    return moved;
}

def inspectUnder(x, y, z) {
    # todo: check for water, lava, etc.
    blocked := [false];
    findShapeUnder(x, y, z, info => {
        if(info[0] = "ground.water") {
            blocked[0] := true;
        }
    });
    return blocked[0] = false;
}

def replaceShape(x, y, z, name) {
    if(intersectsShapes(x, y, z, name, PLAYER_SHAPE) = false) {
        eraseShape(x, y, z);
        setShape(x, y, z, name);
    } else {
        print("player blocks!");
    }
}

def forPlayerBase(fx) {
    range(0, PLAYER_X, 1, x => {
        range(0, PLAYER_Y, 1, y => {
            fx(x, y);
        });
    });
}

def findShapeUnder(px, py, pz, fx) {
    forPlayerBase((x, y) => {
        info := getShape(px + x, py + y, pz - 1);
        if(info != null) {
            fx(info);
        }
    });
}

def findShapeNearby(name, fx) {
    found := [false];
    range(-1, PLAYER_X + 1, 1, x => {
        range(-1, PLAYER_Y + 1, 1, y => {
            range(0, PLAYER_Z, 1, z => {
                if(found[0] = false) {
                    info := getShape(player.x + x, player.y + y, player.z + z);
                    if(info != null) {
                        if(info[0] = name) {
                            fx(info[1], info[2], info[3]);
                            found[0] := true;
                        }
                    }
                }
            });
        });
    });
    return found[0];
}

def getRoofZ() {
    # roofs are only at certain heights
    return (int(player.z / 7) + 1) * 7;
}

def isUnderRoof() {
    found := [true];
    # roofs are only at certain heights
    z := getRoofZ();
    forPlayerBase((x, y) => {
        info := getShape(player.x + x, player.y + y, z);
        if(info = null) {
            found[0] := false;
        } else {
            if(substr(info[0], 0, 5) != "roof.") {
                found[0] := false;
            }
        }
    });
    return found[0];
}

# Put main last so if there are parsing errors, the game panic()-s.
def main() {
    moveViewTo(player.x, player.y);
    setShape(player.x, player.y, player.z, PLAYER_SHAPE);
}
