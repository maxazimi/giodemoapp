package clock

import (
	"bytes"
	_ "embed"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	anim "giowidgets/animation"
	"giowidgets/icon"
	page "giowidgets/pages"
	"image"
	"image/color"
	"math"
	"time"
)

const (
	STEP = 6.0
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	_ page.Page = &Page{}

	Tick         bool
	SpringEasing func(float64) float64

	//go:embed clock_face.png
	clockFace []byte
	//go:embed clock_hour.png
	clockHour []byte
	//go:embed clock_min.png
	clockMin []byte
)

type Page struct {
	*page.Router

	clock struct {
		animateClock            bool
		animStartTime           time.Time
		hourDegrees, minDegrees float64
		deltaDegrees            float64
		faceImg                 paint.ImageOp
		hourImg                 paint.ImageOp
		minImg                  paint.ImageOp
	}
	toggleBtn widget.Clickable
}

func New(router *page.Router) *Page {
	clockPage := &Page{
		Router: router,
	}

	SpringEasing = anim.SpringFactory(anim.Easing{
		HalfCycles:      6,   // accepts integers > 1
		Damping:         0.5, // must be between 0 and 1 excluding 1,
		InitialPosition: -1,  // typically in -1..1 inclusive
		InitialVelocity: 0,   // at start, how fast in either direction
	})

	clockPage.clock.animateClock = false
	clockPage.clock.hourDegrees = 0
	clockPage.clock.minDegrees = 0

	img, _, err := image.Decode(bytes.NewReader(clockFace))
	if err != nil {
		return nil
	}
	clockPage.clock.faceImg = paint.NewImageOp(img)

	img, _, err = image.Decode(bytes.NewReader(clockHour))
	if err != nil {
		return nil
	}
	clockPage.clock.hourImg = paint.NewImageOp(img)

	img, _, err = image.Decode(bytes.NewReader(clockMin))
	if err != nil {
		return nil
	}
	clockPage.clock.minImg = paint.NewImageOp(img)

	return clockPage
}

func (p *Page) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *Page) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *Page) NavItem() component.NavItem {
	return component.NavItem{
		Name: "Clock",
		Icon: icon.ClockIcon,
	}
}

func (p *Page) tick(d float64) {
	p.clock.minDegrees += d
	if p.clock.minDegrees >= 360 {
		p.clock.minDegrees = 0
	}

	p.clock.hourDegrees += d * 30 / 360
	if p.clock.hourDegrees >= 360 {
		p.clock.hourDegrees = 0
	}

	p.clock.animStartTime = time.Now()
	p.clock.animateClock = true
}

func (p *Page) animateClock(gtx C, duration time.Duration, factor float64) {
	elapsed := gtx.Now.Sub(p.clock.animStartTime)
	progress := elapsed.Seconds() / duration.Seconds()

	if progress < 1 {
		op.InvalidateOp{}.Add(gtx.Ops)
	} else {
		progress = 1
		p.clock.animateClock = false
	}
	p.clock.deltaDegrees = factor * SpringEasing(progress)
}

func draw(ops *op.Ops, offset image.Point, origin, scale f32.Point, rotation float64) {
	defer op.Offset(offset).Push(ops).Pop()
	tr := f32.Affine2D{}.Scale(origin, scale)
	if rotation != 0 {
		tr = tr.Rotate(origin, float32(rotation*(math.Pi/180)))
	}
	defer op.Affine(tr).Push(ops).Pop()
	paint.PaintOp{}.Add(ops)
}

func (p *Page) drawClock(ops *op.Ops) D {
	// specify clip area
	maxSize := p.clock.faceImg.Size()
	maxSize.Y -= 250

	const r = 20
	bounds := image.Rectangle{Min: image.Pt(135, 10), Max: maxSize.Sub(image.Pt(135, 10))}
	defer clip.RRect{Rect: bounds, SE: r, SW: r, NW: r, NE: r}.Push(ops).Pop()
	paint.LinearGradientOp{Color1: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
		Color2: color.NRGBA{R: 255, G: 255, B: 255, A: 255}, Stop2: f32.Pt(float32(maxSize.X), float32(maxSize.Y))}.Add(ops)
	paint.PaintOp{}.Add(ops)

	// draw clock face
	p.clock.faceImg.Add(ops)
	draw(ops, image.Pt(0, -125), f32.Pt(300, 300), f32.Pt(0.5, 0.5), 0)

	// draw hour hand
	p.clock.hourImg.Add(ops)
	draw(ops, image.Pt(250, -173), f32.Pt(49, 348), f32.Pt(0.25, 0.25), p.clock.hourDegrees-1)

	// draw minute hand
	p.clock.minImg.Add(ops)
	draw(ops, image.Pt(261, -273), f32.Pt(40, 448), f32.Pt(0.25, 0.25), p.clock.minDegrees+p.clock.deltaDegrees-1)

	return D{Size: maxSize}
}

func (p *Page) Layout(gtx C, th *material.Theme) D {
	if Tick {
		Tick = false
		p.tick(STEP)
	}
	if p.clock.animateClock {
		p.animateClock(gtx, 1000*time.Millisecond, STEP)
	}

	return layout.Flex{Alignment: layout.Middle, Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return p.drawClock(gtx.Ops)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle, Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					if p.toggleBtn.Clicked() {
						// TODO
					}
					return material.Button(th, &p.toggleBtn, "Toggle").Layout(gtx)
				}),
			)
		}),
	)
}
