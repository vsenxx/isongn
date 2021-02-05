package world

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/uzudil/isongn/shapes"
)

const (
	SECTION_SIZE   = 400
	SECTION_Z_SIZE = 32
	VERSION        = 1
)

type Point struct {
	x, y, z int
}

type Section struct {
	X, Y    int
	shapes  [SECTION_SIZE][SECTION_SIZE][SECTION_Z_SIZE]byte
	origins [SECTION_SIZE][SECTION_SIZE][SECTION_Z_SIZE]*Point
}

type SectionCache struct {
	cache [4]*Section
	times [4]int64
}

func NewSectionCache() *SectionCache {
	return &SectionCache{}
}

type Loader struct {
	dir          string
	X, Y         int
	sectionCache *SectionCache
}

func NewLoader(dir string, x, y int) *Loader {
	return &Loader{dir, x, y, NewSectionCache()}
}

func (loader *Loader) SetShape(x, y, z int, shapeIndex byte) bool {
	section, atomX, atomY, atomZ := loader.getPosInSection(x, y, z)
	section.shapes[atomX][atomY][atomZ] = shapeIndex + 1
	// mark the shape's origin
	shapes.Shapes[shapeIndex].Traverse(func(xx, yy, zz int) {
		sec, ax, ay, az := loader.getPosInSection(x+xx, y+yy, z+zz)
		sec.origins[ax][ay][az] = &Point{x, y, z}
	})
	return true
}

func (loader *Loader) EraseShape(x, y, z int) bool {
	section, atomX, atomY, atomZ := loader.getPosInSection(x, y, z)
	if section.origins[atomX][atomY][atomZ] != nil {
		o := section.origins[atomX][atomY][atomZ]
		section, atomX, atomY, atomZ := loader.getPosInSection(o.x, o.y, o.z)
		shapeIndex := section.shapes[atomX][atomY][atomZ]
		if shapeIndex > 0 {
			shapes.Shapes[shapeIndex-1].Traverse(func(xx, yy, zz int) {
				sec, ax, ay, az := loader.getPosInSection(x+xx, y+yy, z+zz)
				sec.origins[ax][ay][az] = nil
			})
			section.shapes[atomX][atomY][atomZ] = 0
			return true
		}
	}
	return false
}

func (loader *Loader) GetShape(worldX, worldY, worldZ int) (byte, int, int, int, bool) {
	section, atomX, atomY, atomZ := loader.getPosInSection(worldX, worldY, worldZ)
	if section.origins[atomX][atomY][atomZ] != nil {
		o := section.origins[atomX][atomY][atomZ]
		section, atomX, atomY, atomZ := loader.getPosInSection(o.x, o.y, o.z)
		shapeIndex := section.shapes[atomX][atomY][atomZ]
		if shapeIndex == 0 {
			return 0, 0, 0, 0, false
		}
		return shapeIndex - 1, o.x, o.y, o.z, true
	}
	return 0, 0, 0, 0, false
}

func (loader *Loader) getPosInSection(worldX, worldY, worldZ int) (*Section, int, int, int) {
	sx := worldX / SECTION_SIZE
	sy := worldY / SECTION_SIZE
	section, err := loader.getSection(sx, sy)
	if err != nil {
		log.Fatal(err)
	}
	atomX := worldX % SECTION_SIZE
	atomY := worldY % SECTION_SIZE
	atomZ := worldZ
	return section, atomX, atomY, atomZ
}

func (loader *Loader) getSection(sx, sy int) (*Section, error) {
	// already loaded?
	oldestIndex := -1
	for i := 0; i < len(loader.sectionCache.cache); i++ {
		if loader.sectionCache.cache[i] != nil && loader.sectionCache.cache[i].X == sx && loader.sectionCache.cache[i].Y == sy {
			loader.sectionCache.times[i] = time.Now().Unix()
			return loader.sectionCache.cache[i], nil
		}
		if oldestIndex == -1 || loader.sectionCache.times[i] < loader.sectionCache.times[oldestIndex] {
			oldestIndex = i
		}
	}

	// save version in cache
	if loader.sectionCache.cache[oldestIndex] != nil {
		err := loader.save(loader.sectionCache.cache[oldestIndex])
		if err != nil {
			return nil, err
		}
	}

	// not found in cache, load it
	section, err := loader.load(sx, sy)
	if err != nil {
		return nil, err
	}

	// put in cache
	loader.sectionCache.cache[oldestIndex] = section
	loader.sectionCache.times[oldestIndex] = time.Now().Unix()

	// expland shapes
	startX := section.X * SECTION_SIZE
	startY := section.Y * SECTION_SIZE
	for x := 0; x < SECTION_SIZE; x++ {
		for y := 0; y < SECTION_SIZE; y++ {
			for z := 0; z < SECTION_Z_SIZE; z++ {
				if section.shapes[x][y][z] > 0 {
					shapes.Shapes[section.shapes[x][y][z]-1].Traverse(func(xx, yy, zz int) {
						sec, ax, ay, az := loader.getPosInSection(startX+x+xx, startY+y+yy, z+zz)
						if sec == section {
							sec.origins[ax][ay][az] = &Point{startX + x, startY + y, z}
						}
					})
				}
			}
		}
	}

	return section, nil
}

func (loader *Loader) SaveAll() error {
	for _, c := range loader.sectionCache.cache {
		if c != nil {
			err := loader.save(c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (loader *Loader) load(sx, sy int) (*Section, error) {
	section := &Section{
		X: sx,
		Y: sy,
	}
	path := loader.getPath(sx, sy)
	fmt.Printf("Looking for section: %s\n", path)
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Loading section: %s\n", path)

		f, err := os.Open(path)
		defer f.Close()
		if err != nil {
			return nil, err
		}

		version := make([]byte, 1)
		_, err = f.Read(version)
		if err != nil {
			return nil, err
		}
		fmt.Printf("\tversion=%d\n", version[0])

		// if version > VERSION... do something
		for x := 0; x < SECTION_SIZE; x++ {
			for y := 0; y < SECTION_SIZE; y++ {
				_, err := f.Read(section.shapes[x][y][:])
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return section, nil
}

func (loader *Loader) save(section *Section) error {
	path := loader.getPath(section.X, section.Y)
	fmt.Printf("Writing section: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	b := []byte{VERSION}
	f.Write(b)
	for x := 0; x < SECTION_SIZE; x++ {
		for y := 0; y < SECTION_SIZE; y++ {
			f.Write(section.shapes[x][y][:])
		}
	}

	return nil
}

func (loader *Loader) getPath(sx, sy int) string {
	return filepath.Join(loader.dir, fmt.Sprintf("map%02x%02x", sx, sy))
}
