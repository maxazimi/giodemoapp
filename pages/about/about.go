package about

import (
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"

	alo "giowidgets/applayout"
	"giowidgets/icon"
	page "giowidgets/pages"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type Page struct {
	aboutBtn widget.Clickable
	widget.List
	*page.Router
}

func New(router *page.Router) *Page {
	return &Page{
		Router: router,
	}
}

var (
	_ page.Page = &Page{}
)

func (p *Page) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *Page) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *Page) NavItem() component.NavItem {
	return component.NavItem{
		Name: "About",
		Icon: icon.OtherIcon,
	}
}

const (
	myURL = "https://github.com/maxazimi"
)

func (p *Page) Layout(gtx C, th *material.Theme) D {
	p.List.Axis = layout.Vertical
	return material.List(th, &p.List).Layout(gtx, 1, func(gtx C, _ int) D {
		return layout.Flex{
			Alignment: layout.Middle,
			Axis:      layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return alo.DefaultInset.Layout(gtx, material.Body1(th, `Nothing yet!`).Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return alo.DetailRow{}.Layout(gtx,
					material.Body1(th, "Take a look at "+myURL).Layout,
					func(gtx C) D {
						if p.aboutBtn.Clicked() {
							clipboard.WriteOp{
								Text: myURL,
							}.Add(gtx.Ops)
						}
						return material.Button(th, &p.aboutBtn, "Copy URL").Layout(gtx)
					})
			}),
		)
	})
}
