package main

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

type CustomWidget struct {
    widget.BaseWidget
    label *widget.Label
}

func NewCustomWidget() *CustomWidget {
    cw := &CustomWidget{
        label: widget.NewLabel("Custom Widget"),
    }
    cw.ExtendBaseWidget(cw)
    return cw
}

func (cw *CustomWidget) CreateRenderer() fyne.WidgetRenderer {
    return &customWidgetRenderer{
        customWidget: cw,
        objects:      []fyne.CanvasObject{cw.label},
    }
}

type customWidgetRenderer struct {
    customWidget *CustomWidget
    objects      []fyne.CanvasObject
}

func (r *customWidgetRenderer) Layout(size fyne.Size) {
    r.customWidget.label.Resize(size)
}

func (r *customWidgetRenderer) MinSize() fyne.Size {
    return r.customWidget.label.MinSize()
}

func (r *customWidgetRenderer) Refresh() {
    r.customWidget.Refresh()
}

func (r *customWidgetRenderer) Objects() []fyne.CanvasObject {
    return r.objects
}

func (r *customWidgetRenderer) Destroy() {
    // Clean up resources if necessary
}
