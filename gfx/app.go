package gfx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/isongn/shapes"
	"github.com/uzudil/isongn/world"
)

type Game interface {
	Init(app *App, config map[string]interface{})
	Name() string
	Events(delta float64)
	GetZ() int
}

type KeyPress struct {
	Key      glfw.Key
	Scancode int
	Action   glfw.Action
	Mods     glfw.ModifierKey
	First    bool
}

type AppConfig struct {
	GameDir    string
	Title      string
	Name       string
	Version    float64
	Font       string
	ViewSize   int
	ViewSizeZ  int
	SectorSize int
	runtime    map[string]interface{}
	zoom       float64
	camera     [3]float32
	shear      [3]float32
	shapes     []map[string]interface{}
	creatures  []map[string]interface{}
}

type App struct {
	Game                            Game
	Font                            *Font
	Config                          *AppConfig
	Window                          *glfw.Window
	KeyState                        map[glfw.Key]*KeyPress
	targetFps                       float64
	lastUpdate                      float64
	nbFrames                        int
	View                            *View
	Ui                              *Ui
	Dir                             string
	Loader                          *world.Loader
	Width, Height                   int
	windowWidth, windowHeight       int
	windowWidthDpi, windowHeightDpi int
	dpiX, dpiY                      float32
	frameBuffer, uiFrameBuffer      *FrameBuffer
}

func NewApp(game Game, gameDir string, windowWidth, windowHeight int, targetFps float64) *App {
	appConfig := parseConfig(gameDir)
	width, height := getResolution(appConfig, game.Name())
	app := &App{
		Game:         game,
		Config:       appConfig,
		KeyState:     map[glfw.Key]*KeyPress{},
		targetFps:    targetFps,
		Width:        width,
		Height:       height,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
	}
	font, err := NewFont(filepath.Join(gameDir, appConfig.Font), 32)
	if err != nil {
		panic(err)
	}
	app.Font = font
	app.Dir = initUserdir(appConfig.Name)
	app.Window = initWindow(windowWidth, windowHeight)
	pxWidth, pxHeight := app.Window.GetFramebufferSize()
	app.dpiX = float32(pxWidth) / float32(windowWidth)
	app.dpiY = float32(pxHeight) / float32(windowHeight)
	fmt.Printf("Resolution: %dx%d Window: %dx%d Dpi: %fx%f\n", app.Width, app.Height, windowWidth, windowHeight, app.dpiX, app.dpiY)
	app.windowWidthDpi = int(float32(app.windowWidth) * app.dpiX)
	app.windowHeightDpi = int(float32(app.windowHeight) * app.dpiY)
	app.Window.SetKeyCallback(app.Keypressed)
	app.Window.SetScrollCallback(app.MouseScroll)
	app.frameBuffer = NewFrameBuffer(int32(width), int32(height), true)
	app.uiFrameBuffer = NewFrameBuffer(int32(width), int32(height), false)
	InitScript()
	err = shapes.InitShapes(gameDir, appConfig.shapes)
	if err != nil {
		panic(err)
	}
	err = shapes.InitCreatures(gameDir, appConfig.creatures)
	if err != nil {
		panic(err)
	}
	app.Loader = world.NewLoader(app.Dir, 1000, 1000)
	app.View = InitView(appConfig.zoom, appConfig.camera, appConfig.shear, app.Loader)
	app.Ui = InitUi(width, height)
	return app
}

func parseConfig(gameDir string) *AppConfig {
	configPath := filepath.Join(gameDir, "config.json")
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	view := data["view"].(map[string]interface{})
	camera := view["camera"].([]interface{})
	shear := view["shear"].([]interface{})
	config := &AppConfig{
		GameDir:    gameDir,
		Title:      data["title"].(string),
		Name:       strings.ToLower(data["name"].(string)),
		Version:    data["version"].(float64),
		Font:       data["font"].(string),
		ViewSize:   int(view["size"].(float64)),
		ViewSizeZ:  int(view["sizeZ"].(float64)),
		SectorSize: int(view["sector"].(float64)),
		runtime:    data["runtime"].(map[string]interface{}),
		zoom:       view["zoom"].(float64),
		camera:     [3]float32{float32(camera[0].(float64)), float32(camera[1].(float64)), float32(camera[2].(float64))},
		shear:      [3]float32{float32(shear[0].(float64)), float32(shear[1].(float64)), float32(shear[2].(float64))},
		shapes:     toMap(data["shapes"].([]interface{})),
		creatures:  toMap(data["creatures"].([]interface{})),
	}
	fmt.Printf("Starting game: %s (v%f)\n", config.Title, config.Version)
	return config
}

func toMap(a []interface{}) []map[string]interface{} {
	r := []map[string]interface{}{}
	for _, o := range a {
		r = append(r, o.(map[string]interface{}))
	}
	return r
}

func getResolution(appConfig *AppConfig, mode string) (int, int) {
	runtimeConfig, ok := appConfig.runtime[mode]
	if ok == false {
		panic("Can't find runtime config")
	}
	resolution, ok := (runtimeConfig.(map[string]interface{}))["resolution"]
	if ok == false {
		panic("Can't find resolution in runtime config")
	}
	resArray := (resolution.([]interface{}))
	return int(resArray[0].(float64)), int(resArray[1].(float64))
}

func initUserdir(gameName string) string {
	// create user dir if needed
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join(userHomeDir, "."+gameName)
	fmt.Printf("Game state path: %s\n", dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
	return dir
}

func initWindow(windowWidth, windowHeight int) *glfw.Window {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "isongn", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	return window
}

func (app *App) IsDown(key glfw.Key) bool {
	_, ok := app.KeyState[key]
	return ok
}

func (app *App) IsDownMod(key glfw.Key, mod glfw.ModifierKey) bool {
	event, ok := app.KeyState[key]
	if ok {
		return event.Mods&mod > 0
	}
	return false
}

func (app *App) IsFirstDown(key glfw.Key) bool {
	event, ok := app.KeyState[key]
	if ok && event.First {
		event.First = false
		return true
	}
	return false
}

func (app *App) Keypressed(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Release {
		delete(app.KeyState, key)
	} else {
		event, ok := app.KeyState[key]
		if ok {
			event.First = false
		} else {
			event = &KeyPress{
				First: true,
			}
		}
		event.Key = key
		event.Scancode = scancode
		event.Action = action
		event.Mods = mods
		app.KeyState[key] = event
	}
}

func (app *App) IsDownAlt1(key1 glfw.Key) bool {
	return app.IsFirstDown(key1) || app.IsDownMod(key1, glfw.ModShift)
}

func (app *App) IsDownAlt(key1, key2 glfw.Key) bool {
	return app.IsDownAlt1(key1) || app.IsDownAlt1(key2)
}

func (app *App) MouseScroll(w *glfw.Window, xoffs, yoffs float64) {
	app.View.Zoom(yoffs)
}

func (app *App) CalcFps() {
	currentTime := glfw.GetTime()
	delta := currentTime - app.lastUpdate
	app.nbFrames++
	if delta >= 1.0 { // If last cout was more than 1 sec ago
		app.Window.SetTitle(fmt.Sprintf("%s - %.2f", app.Config.Title, float64(app.nbFrames)/delta))
		app.nbFrames = 0
		app.lastUpdate = currentTime
	}
}

func (app *App) Sleep(lastTime float64) (float64, float64) {
	now := glfw.GetTime()
	d := now - lastTime
	sleep := ((1.0 / app.targetFps) - d) * 1000.0
	if sleep > 0 {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}
	return now, d
}

func (app *App) Run() {
	app.Game.Init(app, app.Config.runtime[app.Game.Name()].(map[string]interface{}))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	// gl.ClearColor(0, 0, 0, 0)

	last := glfw.GetTime()
	var delta float64
	for !app.Window.ShouldClose() {
		// reduce fan noise / run at target fps
		last, delta = app.Sleep(last)

		// show FPS in window title
		app.CalcFps()

		// handle events
		app.Game.Events(delta)

		app.frameBuffer.Enable(app.Width, app.Height)
		app.View.Draw(delta)
		app.frameBuffer.Draw(app.windowWidthDpi, app.windowHeightDpi)

		app.uiFrameBuffer.Enable(app.Width, app.Height)
		app.Ui.Draw()
		app.uiFrameBuffer.Draw(app.windowWidthDpi, app.windowHeightDpi)

		// Maintenance
		app.Window.SwapBuffers()
		glfw.PollEvents()
	}
}
