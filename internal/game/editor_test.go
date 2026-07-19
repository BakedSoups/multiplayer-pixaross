package game

import (
	"encoding/json"
	"testing"
)

func TestEditorPaintingBuildsSolutionAutomatically(t *testing.T) {
	editor := newEditorState(4)
	editor.apply(1, 2)

	cell := editor.Cells[editor.index(1, 2)]
	if !cell.Visible || !cell.Filled {
		t.Fatalf("painted cell = %+v, want visible and filled", cell)
	}

	editor.Tool = editorToolEraser
	editor.apply(1, 2)
	cell = editor.Cells[editor.index(1, 2)]
	if cell.Visible || cell.Filled {
		t.Fatalf("erased cell = %+v, want hidden and unfilled", cell)
	}
}

func TestEditorLineConnectsFastPointerMovement(t *testing.T) {
	editor := newEditorState(6)
	editor.applyLine(0, 0, 5, 5)

	for i := 0; i < 6; i++ {
		cell := editor.Cells[editor.index(i, i)]
		if !cell.Visible || !cell.Filled {
			t.Fatalf("line cell (%d,%d) = %+v, want visible and filled", i, i, cell)
		}
	}
}

func TestEditorBeforeAndAfterLayersStayIndependent(t *testing.T) {
	editor := newEditorState(3)
	editor.selectLayer(editorLayerBefore)
	editor.PaintColor.R = 255
	editor.apply(1, 1)

	cell := editor.Cells[editor.index(1, 1)]
	if !cell.BeforeVisible || cell.Visible || cell.Filled {
		t.Fatalf("before paint changed the after layer: %+v", cell)
	}
	if !sameRGBA(cell.BeforeColor, editorBeforeColor) {
		t.Fatalf("before color = %+v, want black", cell.BeforeColor)
	}

	editor.selectLayer(editorLayerAfter)
	editor.apply(2, 1)
	cell = editor.Cells[editor.index(2, 1)]
	if !cell.Visible || !cell.Filled || cell.BeforeVisible {
		t.Fatalf("after paint changed the before layer: %+v", cell)
	}

	puzzle := editor.puzzle()
	if puzzle.SkeletonRaw[1][1] == "" {
		t.Fatal("before layer was not exported as skeleton art")
	}
	if puzzle.RevealRaw[1][2] == "" {
		t.Fatal("after layer was not exported as reveal art")
	}
}

func TestParseEditorBrowserColor(t *testing.T) {
	c, ok := parseEditorHexColor("#12A4EF")
	if !ok || c.R != 0x12 || c.G != 0xA4 || c.B != 0xEF || c.A != 0xFF {
		t.Fatalf("parsed color = %+v, %v", c, ok)
	}
}

func TestEditorImageExportUsesAfterLayer(t *testing.T) {
	editor := newEditorState(2)
	editor.selectLayer(editorLayerBefore)
	editor.apply(0, 0)
	editor.selectLayer(editorLayerAfter)
	editor.apply(1, 0)

	var payload editorImportPayload
	if err := json.Unmarshal([]byte(editor.imageExportJSON()), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Cells[0].A != 0 {
		t.Fatalf("before-only pixel was exported: %+v", payload.Cells[0])
	}
	if payload.Cells[1].A == 0 {
		t.Fatalf("after pixel was not exported: %+v", payload.Cells[1])
	}
}

func TestEditorResizePreservesArt(t *testing.T) {
	editor := newEditorState(10)
	editor.selectLayer(editorLayerBefore)
	editor.apply(2, 2)
	editor.selectLayer(editorLayerAfter)
	editor.apply(5, 5)

	resized := editor.resized(32)
	before := resized.Cells[resized.index(7, 7)]
	if !before.BeforeVisible {
		t.Fatal("before art disappeared after resize")
	}
	after := resized.Cells[resized.index(16, 16)]
	if !after.Visible || !after.Filled {
		t.Fatal("after art disappeared after resize")
	}
}

func TestEditorPreviewReturnsToEditor(t *testing.T) {
	game := Game{mode: screenPuzzle, editorPreview: true}
	game.leavePuzzle()
	if game.mode != screenEditor || game.editorPreview {
		t.Fatalf("leave puzzle = mode %v, preview %v", game.mode, game.editorPreview)
	}

	game.mode = screenReveal
	game.editorPreview = true
	game.leaveReveal()
	if game.mode != screenEditor || game.editorPreview {
		t.Fatalf("leave reveal = mode %v, preview %v", game.mode, game.editorPreview)
	}
}
