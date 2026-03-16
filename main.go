// Package main is the entry point for the Ashta Changa game.
// This file is minimal and delegates to the gui package for the game implementation.
package main

import (
	"log"

	"github.com/anujdecoder/ashta-board/gui"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := gui.NewGame()

	ebiten.SetWindowSize(gui.ScreenWidth, gui.ScreenHeight)
	ebiten.SetWindowTitle("Ashta Changa")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
