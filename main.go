// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Renders a textured spinning cube using GLFW 3 and OpenGL 4.1 core forward-compatible profile.
package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/uzudil/isongn/editor"
	"github.com/uzudil/isongn/gfx"
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	game := flag.String("game", ".", "Location of the game assets directory")
	width := flag.Int("width", 320, "Screen width (default: 320)")
	height := flag.Int("height", 240, "Screen height (default: 200)")
	winWidth := flag.Int("winwidth", 800, "Window width (default: 800)")
	winHeight := flag.Int("winheight", 600, "Window height (default: 600)")
	fps := flag.Float64("fps", 60, "Frames per second")
	flag.Parse()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	app := gfx.NewApp(*game, *width, *height, *winWidth, *winHeight, *fps)
	app.Run(editor.NewEditor(app))
}
