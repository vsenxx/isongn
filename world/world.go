package world

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"unsafe"
)

const (
	SECTION_SIZE   = 200
	SECTION_Z_SIZE = 24
	STAMP_SIZE     = 8
	VERSION        = 1
)

type Point struct {
	x, y, z int
}

type Position struct {
	Shape int
}

type Section struct {
	X, Y     int
	position [SECTION_SIZE][SECTION_SIZE][SECTION_Z_SIZE]Position
	edges    [SECTION_SIZE][SECTION_SIZE]Position
}

type SectionCache struct {
	cache [4]*Section
	times [4]int64
}

func NewSectionCache() *SectionCache {
	return &SectionCache{}
}

type Display interface {
	Invalidate()
}

type Loader struct {
	dir          string
	X, Y         int
	sectionCache *SectionCache
	display      Display
}

func NewLoader(dir string, x, y int, display Display) *Loader {
	return &Loader{dir, x, y, NewSectionCache(), display}
}

func (loader *Loader) MoveTo(x, y int) bool {
	if x >= 0 && y >= 0 && (loader.X != x || loader.Y != y) {
		loader.X = x
		loader.Y = y
		loader.display.Invalidate()
		return true
	}
	return false
}

func (loader *Loader) ClearEdge(x, y int) {
	section, atomX, atomY, _ := loader.getPosInSection(x, y, 0)
	section.edges[atomX][atomY].Shape = 0
	// loader.display.Invalidate()
}

func (loader *Loader) SetEdge(x, y int, shapeIndex int) {
	section, atomX, atomY, _ := loader.getPosInSection(x, y, 0)
	section.edges[atomX][atomY].Shape = shapeIndex + 1
	// loader.display.Invalidate()
}

func (loader *Loader) GetEdge(x, y int) (int, bool) {
	section, atomX, atomY, _ := loader.getPosInSection(x, y, 0)
	shapeIndex := section.edges[atomX][atomY].Shape
	if shapeIndex == 0 {
		return 0, false
	}
	return shapeIndex - 1, true
}

func (loader *Loader) SetShape(x, y, z int, shapeIndex int) bool {
	section, atomX, atomY, atomZ := loader.getPosInSection(x, y, z)
	section.position[atomX][atomY][atomZ].Shape = shapeIndex + 1
	// loader.display.Invalidate()
	return true
}

func (loader *Loader) EraseShape(x, y, z int) bool {
	section, atomX, atomY, atomZ := loader.getPosInSection(x, y, z)
	shapeIndex := section.position[atomX][atomY][atomZ].Shape
	if shapeIndex > 0 {
		section.position[atomX][atomY][atomZ].Shape = 0
		// loader.display.Invalidate()
		return true
	}
	return false
}

func (loader *Loader) GetShape(worldX, worldY, worldZ int) (int, bool) {
	section, atomX, atomY, atomZ := loader.getPosInSection(worldX, worldY, worldZ)
	shapeIndex := section.position[atomX][atomY][atomZ].Shape
	if shapeIndex == 0 {
		return 0, false
	}
	return shapeIndex - 1, true
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
		defer un(trace("Loading map"))
		fmt.Printf("Loading section: %s bytes: %d\n", path, unsafe.Sizeof(section.position))

		f, err := os.Open(path)
		defer f.Close()
		if err != nil {
			return nil, err
		}

		fz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer fz.Close()

		version := make([]byte, 1)
		_, err = fz.Read(version)
		if err != nil {
			return nil, err
		}
		fmt.Printf("\tversion=%d\n", version[0])

		// if version > VERSION... do something
		dec := gob.NewDecoder(fz)
		err = dec.Decode(&section.position)
		if err != nil {
			return nil, err
		}
		err = dec.Decode(&section.edges)
		if err != nil {
			return nil, err
		}
	}
	return section, nil
}

func (loader *Loader) save(section *Section) error {
	defer un(trace("Saving map"))
	path := loader.getPath(section.X, section.Y)
	fmt.Printf("Writing section: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fz := gzip.NewWriter(f)
	defer fz.Close()

	b := []byte{VERSION}
	fz.Write(b)
	enc := gob.NewEncoder(fz)
	err = enc.Encode(section.position)
	if err != nil {
		return err
	}
	err = enc.Encode(section.edges)
	if err != nil {
		return err
	}

	return nil
}

func (loader *Loader) getPath(sx, sy int) string {
	return filepath.Join(loader.dir, fmt.Sprintf("map%02x%02x", sx, sy))
}

func trace(s string) (string, time.Time) {
	log.Println("START:", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("  END:", s, "ElapsedTime in seconds:", endTime.Sub(startTime))
}
