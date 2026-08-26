package main

import (
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"boxel/internal/app"
	"boxel/internal/hub"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "serve" || os.Args[1] == "server") {
		log.Fatal(hub.ListenAndServe(os.Getenv("HOST"), os.Getenv("PORT")))
	}

	go func() {
		_ = hub.ListenAndServe("127.0.0.1", "8787")
	}()
	time.Sleep(80 * time.Millisecond)

	ebiten.SetWindowTitle("BOXEL")
	ebiten.SetWindowSize(1440, 900)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(app.New()); err != nil {
		log.Fatal(err)
	}
}
