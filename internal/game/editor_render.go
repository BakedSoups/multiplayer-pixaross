package game

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) drawEditor(screen *ebiten.Image) {
	screen.Fill(colPanel)
	heading := "DRAW"
	if g.editingProfile {
		heading = "PROFILE"
	}
	drawScaledTextCentered(screen, heading, rect{x: 150, y: 22, w: 240, h: 42}, 1.85, colInk)
	drawButton(screen, editorBackButton(), "back")
	drawButton(screen, editorUndoButton(), "undo")
	if !g.editingProfile {
		drawButton(screen, editorSizeButton(), fmt.Sprintf("%dx%d v", g.editor.Width, g.editor.Height))
		title := g.editor.Title
		if len(title) > 18 {
			title = title[:18]
		}
		drawButton(screen, editorTitleButton(), title)
	}

	g.drawEditorGrid(screen)
	g.drawEditorToolbar(screen)
	g.drawEditorPalette(screen)
	if g.editingProfile {
		drawButton(screen, profileSaveButton(), "save profile")
	} else {
		g.drawEditorLayers(screen)
		g.drawEditorActions(screen)
	}
	if g.editorSizeOpen && !g.editingProfile {
		g.drawEditorSizeMenu(screen)
	}

	if time.Now().Before(g.menuNoticeUntil) {
		drawCenteredText(screen, g.menuNotice, rect{x: 120, y: 64, w: 300, h: 24}, colAccent)
	}
}

func (g *Game) drawEditorGrid(screen *ebiten.Image) {
	grid := editorGridRect(g.editor)
	drawRounded(screen, rect{x: grid.x - 8, y: grid.y - 8, w: grid.w + 16, h: grid.h + 16}, 8, colGridHeavy)
	vector.DrawFilledRect(screen, float32(grid.x), float32(grid.y), float32(grid.w), float32(grid.h), colWhite, false)
	cell := editorCellSize(g.editor)
	for y := 0; y < g.editor.Height; y++ {
		for x := 0; x < g.editor.Width; x++ {
			r := rect{x: grid.x + float64(x)*cell, y: grid.y + float64(y)*cell, w: cell, h: cell}
			if (x+y)%2 == 1 {
				vector.DrawFilledRect(screen, float32(r.x), float32(r.y), float32(r.w), float32(r.h), color.RGBA{244, 239, 224, 255}, false)
			}
			cellState := g.editor.Cells[g.editor.index(x, y)]
			if g.editor.Layer == editorLayerAfter && g.editorOnionSkin {
				beforeColor, beforeVisible := g.editor.pixelForLayer(cellState, editorLayerBefore)
				if beforeVisible {
					vector.DrawFilledRect(screen, float32(r.x), float32(r.y), float32(r.w), float32(r.h), alphaColor(beforeColor, 0.24), false)
				}
			}
			c, visible := g.editor.pixelForLayer(cellState, g.editor.Layer)
			if visible {
				vector.DrawFilledRect(screen, float32(r.x), float32(r.y), float32(r.w), float32(r.h), c, false)
			}
		}
	}
	for x := 0; x <= g.editor.Width; x++ {
		thick := float32(1)
		line := colGrid
		if x%5 == 0 {
			thick = 2
			line = colGridHeavy
		}
		xx := float32(grid.x + float64(x)*cell)
		vector.StrokeLine(screen, xx, float32(grid.y), xx, float32(grid.y+cell*float64(g.editor.Height)), thick, line, false)
	}
	for y := 0; y <= g.editor.Height; y++ {
		thick := float32(1)
		line := colGrid
		if y%5 == 0 {
			thick = 2
			line = colGridHeavy
		}
		yy := float32(grid.y + float64(y)*cell)
		vector.StrokeLine(screen, float32(grid.x), yy, float32(grid.x+cell*float64(g.editor.Width)), yy, thick, line, false)
	}
}

func (g *Game) drawEditorToolbar(screen *ebiten.Image) {
	drawEditorToolButton(screen, editorPencilButton(), g.editor.Tool == editorToolPencil)
	drawPixelIconImage(screen, g.icons.Pencil, editorPencilButton())
	drawEditorToolButton(screen, editorEraserButton(), g.editor.Tool == editorToolEraser)
	drawPixelIconImage(screen, g.icons.Eraser, editorEraserButton())
	drawEditorToolButton(screen, editorFillButton(), g.editor.Tool == editorToolFill)
	drawPixelBucketIcon(screen, editorFillButton(), g.editor.Tool == editorToolFill)
	drawEditorToolButton(screen, editorEyeButton(), g.editor.Tool == editorToolEyedropper)
	drawPixelIconImage(screen, g.icons.Eyedropper, editorEyeButton())
}

func drawPixelBucketIcon(screen *ebiten.Image, r rect, active bool) {
	ink := color.Color(colInk)
	paint := color.Color(colBlue)
	if active {
		ink = colWhite
		paint = colAccentSoft
	}
	x := float32(r.x + 20)
	y := float32(r.y + 11)
	vector.StrokeLine(screen, x+3, y+3, x+20, y+20, 4, ink, false)
	vector.StrokeLine(screen, x+20, y+20, x+10, y+30, 4, ink, false)
	vector.StrokeLine(screen, x+10, y+30, x-7, y+13, 4, ink, false)
	vector.StrokeLine(screen, x-7, y+13, x+3, y+3, 4, ink, false)
	vector.DrawFilledRect(screen, x+3, y+23, 16, 5, paint, false)
	vector.DrawFilledRect(screen, x+23, y+27, 6, 6, paint, false)
}

func (g *Game) drawEditorPalette(screen *ebiten.Image) {
	if g.editor.Layer == editorLayerBefore {
		r := editorPaletteRect(0)
		drawRounded(screen, r, 5, editorBeforeColor)
		drawRectOutline(screen, rect{x: r.x - 3, y: r.y - 3, w: r.w + 6, h: r.h + 6}, 3, colInk)
		drawCenteredText(screen, "BLACK ONLY", rect{x: 110, y: 603, w: 330, h: 36}, colMuted)
		return
	}
	for i, c := range editorPalette {
		r := editorPaletteRect(i)
		drawRounded(screen, r, 5, c)
		if sameRGBA(c, g.editor.PaintColor) {
			drawRectOutline(screen, rect{x: r.x - 3, y: r.y - 3, w: r.w + 6, h: r.h + 6}, 3, colInk)
		}
	}
	drawRainbowSwatch(screen, editorRainbowRect())
	if !isEditorPaletteColor(g.editor.PaintColor) {
		drawRectOutline(screen, inset(editorRainbowRect(), -3), 3, colInk)
	}
}

func (g *Game) drawEditorLayers(screen *ebiten.Image) {
	drawText(screen, "layer", 44, 674, colMuted)
	drawButton(screen, editorBeforeButton(), modeLabel("before", g.editor.Layer == editorLayerBefore))
	drawButton(screen, editorAfterButton(), modeLabel("after", g.editor.Layer == editorLayerAfter))
	if g.editor.Layer == editorLayerAfter {
		drawEditorToolButton(screen, editorLayerPreviewButton(), g.editorOnionSkin)
		drawPixelIconImageSized(screen, g.icons.Eye, editorLayerPreviewButton(), 24)
	}
	drawText(screen, editorColorHex(g.editor.PaintColor), 424, 674, colMuted)
}

func (g *Game) drawEditorActions(screen *ebiten.Image) {
	drawButton(screen, editorSaveButton(), "save")
	drawButton(screen, editorExportButton(), "export")
	drawButton(screen, editorPreviewButton(), "play")
}

func (g *Game) drawEditorSizeMenu(screen *ebiten.Image) {
	for _, size := range editorSizes {
		drawButton(screen, editorSizeOption(size), modeLabel(fmt.Sprintf("%dx%d", size, size), g.editor.Width == size && g.editor.Height == size))
	}
}

func drawEditorToolButton(screen *ebiten.Image, r rect, active bool) {
	drawIconButton(screen, r)
	if active {
		drawRectOutline(screen, inset(r, 3), 3, colAccent)
	}
}

func drawPixelIconImage(dst, icon *ebiten.Image, r rect) {
	drawPixelIconImageSized(dst, icon, r, 34)
}

func drawPixelIconImageSized(dst, icon *ebiten.Image, r rect, size float64) {
	b := imageBounds(icon)
	scale := math.Min(size/float64(b.Dx()), size/float64(b.Dy()))
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(r.x+(r.w-float64(b.Dx())*scale)/2, r.y+(r.h-float64(b.Dy())*scale)/2)
	dst.DrawImage(icon, op)
}

func drawRainbowSwatch(screen *ebiten.Image, r rect) {
	colors := []color.RGBA{
		{226, 76, 70, 255}, {239, 150, 55, 255}, {239, 207, 74, 255},
		{76, 170, 92, 255}, {68, 142, 210, 255}, {115, 91, 188, 255},
	}
	drawRounded(screen, r, 5, colWhite)
	inner := inset(r, 3)
	stripe := inner.w / float64(len(colors))
	for i, c := range colors {
		vector.DrawFilledRect(screen, float32(inner.x+float64(i)*stripe), float32(inner.y), float32(math.Ceil(stripe)), float32(inner.h), c, false)
	}
	drawRectOutline(screen, r, 2, colGridHeavy)
}

func isEditorPaletteColor(c color.RGBA) bool {
	for _, candidate := range editorPalette {
		if sameRGBA(c, candidate) {
			return true
		}
	}
	return false
}

func modeLabel(label string, active bool) string {
	if active {
		return "[" + label + "]"
	}
	return label
}

func editorGridArea() rect { return rect{x: 40, y: 98, w: 460, h: 430} }

func editorGridRect(e editorState) rect {
	area := editorGridArea()
	cell := editorCellSize(e)
	w := cell * float64(e.Width)
	h := cell * float64(e.Height)
	return rect{x: area.x + (area.w-w)/2, y: area.y + (area.h-h)/2, w: w, h: h}
}

func editorCellSize(e editorState) float64 {
	area := editorGridArea()
	return float64(int(minFloat(area.w/float64(e.Width), area.h/float64(e.Height))))
}

func editorCellAt(e editorState, px, py int) (int, int, bool) {
	grid := editorGridRect(e)
	if float64(px) < grid.x || float64(px) >= grid.x+grid.w || float64(py) < grid.y || float64(py) >= grid.y+grid.h {
		return 0, 0, false
	}
	cell := editorCellSize(e)
	x := int((float64(px) - grid.x) / cell)
	y := int((float64(py) - grid.y) / cell)
	return x, y, e.inBounds(x, y)
}

func editorBackButton() rect   { return rect{x: 24, y: 24, w: 82, h: 38} }
func editorUndoButton() rect   { return rect{x: 116, y: 24, w: 82, h: 38} }
func editorSizeButton() rect   { return rect{x: 404, y: 24, w: 112, h: 38} }
func editorTitleButton() rect  { return rect{x: 176, y: 68, w: 188, h: 26} }
func editorPencilButton() rect { return rect{x: 108, y: 542, w: 64, h: 48} }
func editorEraserButton() rect { return rect{x: 195, y: 542, w: 64, h: 48} }
func editorFillButton() rect   { return rect{x: 282, y: 542, w: 64, h: 48} }
func editorEyeButton() rect    { return rect{x: 369, y: 542, w: 64, h: 48} }
func editorBeforeButton() rect { return rect{x: 130, y: 653, w: 118, h: 34} }
func editorAfterButton() rect  { return rect{x: 258, y: 653, w: 118, h: 34} }
func editorLayerPreviewButton() rect {
	return rect{x: 382, y: 653, w: 36, h: 34}
}
func editorSaveButton() rect    { return rect{x: 60, y: 700, w: 130, h: 38} }
func editorExportButton() rect  { return rect{x: 205, y: 700, w: 130, h: 38} }
func editorPreviewButton() rect { return rect{x: 350, y: 700, w: 130, h: 38} }
func profileSaveButton() rect   { return rect{x: 174, y: 680, w: 192, h: 44} }
func editorPaletteRect(index int) rect {
	return rect{x: 48 + float64(index)*56, y: 603, w: 36, h: 36}
}

func editorRainbowRect() rect { return editorPaletteRect(len(editorPalette)) }

var editorSizes = []int{8, 10, 15, 20}

func editorSizeOption(size int) rect {
	index := 0
	for i, candidate := range editorSizes {
		if candidate == size {
			index = i
			break
		}
	}
	return rect{x: 404, y: 68 + float64(index)*38, w: 112, h: 34}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
