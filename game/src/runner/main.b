# the global player state
player := {
    "x": 0,
    "y": 0,
    "z": 0,
    "underRoof": false,
};

# the player's shape size
const PLAYER_X = 2;
const PLAYER_Y = 2;
const PLAYER_Z = 4;

# Called when the player steps on a specific shape.
# Return false if it's a shape we should not allow the player to step on.
# Note that in order for this function to be called, the shape name should first be registered (in main).
def onShape(name) {
    if(name = "ground.water") {
        return false;
    }
    return true;
}

# called when the player moves
def onPlayerMove(x, y, z) {
    player.x := x;
    player.y := y;
    player.z := z;
    player.underRoof := isUnderRoof();

    # if under a roof, hide the roof
    if(player.underRoof) {
        setMaxZ(getRoofZ());
    } else {
        setMaxZ(24);
    }
}

# Called for keys we registered in main.
def onKey(key) {
    if(key = KeySpace) {
        if(findShapeNearby("door.wood.y", (x,y,z) => replaceShape(x, y, z, "door.wood.x")) = false) {
            if(findShapeNearby("door.wood.x", (x,y,z) => replaceShape(x, y, z, "door.wood.y")) = false) {
                # do something else
            }
        }
    }
}

def replaceShape(x, y, z, name) {
    if(intersectsPlayer(x, y, z, name) = false) {
        eraseShape(x, y, z);
        setShape(x, y, z, name);
    } else {
        print("player blocks!");
    }
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
    range(0, PLAYER_X, 1, x => {
        range(0, PLAYER_Y, 1, y => {
            info := getShape(player.x + x, player.y + y, z);
            if(info = null) {
                found[0] := false;
            } else {
                if(substr(info[0], 0, 5) != "roof.") {
                    found[0] := false;
                }
            }
        });
    });
    return found[0];
}

# Put main last so if there are parsing errors, the game panic()-s.
def main() {
    # register shape names we are interested in (eg. blocking shapes)
    registerShape("ground.water");

    # register keys we want called back
    registerKey(KeySpace);
}
