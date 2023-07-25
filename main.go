package main

import (
	"flag"
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	page "giowidgets/pages"
	"giowidgets/pages/about"
	"giowidgets/pages/clock"
	"log"
	"os"
	"time"
)

var (
	tick chan bool
)

func main() {
	flag.Parse()

	tick = make(chan bool)

	go func() {
		for {
			tick <- true
			time.Sleep(time.Second)
		}
	}()

	go func() {
		w := app.NewWindow(
			app.Title("MyGioApp"),
			app.Size(unit.Dp(600), unit.Dp(500)),
		)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops

	router := page.NewRouter()
	router.Register(1, clock.New(&router))
	router.Register(2, about.New(&router))

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				router.Layout(gtx, th)
				e.Frame(gtx.Ops)
			}
		case p := <-tick:
			clock.Tick = p
			w.Invalidate()
		}
	}
}
