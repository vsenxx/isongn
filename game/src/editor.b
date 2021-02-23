const TREES = [ "plant.oak", "plant.red", "plant.pine", "plant.willow", "plant.dead" ];

def choose(a) {
    return a[int(random() * len(a))];
}

def editorCommand() {
    if(isPressed(KeyT)) {
        pos := getPosition();
        setShape(pos[0], pos[1], pos[2], "plant.trunk");
        setShape(pos[0] - 1, pos[1] - 1, pos[2] + 4, choose(TREES));
    }
}

def main() {
    # this is never called
}
