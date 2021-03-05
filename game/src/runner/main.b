def main() {
    # register shape names we are interested in (eg. blocking shapes)
    registerShape("ground.water");
}

# Called when the player steps on a specific shape.
# Return false if it's a shape we should not allow the player to step on.
# Note that in order for this function to be called, the shape name should first be registered (in main).
def onShape(name) {
    if(name = "ground.water") {
        return false;
    }
    return true;
}

