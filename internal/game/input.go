package game

import (
	"image"
	"strings"
	"time"

	"github.com/alex/nongrampictures/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) updateInput() {
	if g.mode == screenMainMenu {
		g.updateMainMenuInput()
		return
	}
	if g.mode == screenLevelSelect {
		g.updateLevelSelectInput()
		return
	}
	if g.mode == screenSettings {
		g.updateSettingsInput()
		return
	}
	if g.mode == screenEditor {
		g.updateEditorInput()
		return
	}
	if g.mode == screenCommunity {
		g.updateCommunityInput()
		return
	}
	if g.mode == screenReveal {
		g.updateRevealInput()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.tool = nonogram.ToolFill
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) || inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.tool = nonogram.ToolMark
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.leavePuzzle()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.godModeFill()
		return
	}

	x, y, down, justPressed, justReleased := pointerState()
	if justReleased {
		g.pointerDown = false
		g.dragging = false
		g.lastCellX = -1
		g.lastCellY = -1
		g.strokeState = nonogram.CellEmpty
	}
	if !down {
		return
	}

	if justPressed {
		switch {
		case g.layout.fillTrigger.Contains(x, y):
			g.tool = nonogram.ToolFill
			return
		case g.layout.markTrigger.Contains(x, y):
			g.tool = nonogram.ToolMark
			return
		case g.layout.godModeButton.Contains(x, y):
			g.godModeFill()
			return
		case g.layout.menuButton.Contains(x, y):
			g.leavePuzzle()
			return
		case g.layout.settingsButton.Contains(x, y):
			g.mode = screenSettings
			return
		}
	}

	cellX, cellY, ok := g.layout.CellAt(x, y, g.board.Width, g.board.Height)
	if !ok {
		return
	}
	if justPressed {
		g.pushUndo()
		g.pointerDown = true
		g.strokeState = nonogram.TargetState(g.tool)
		if g.board.Cells[cellY][cellX] == g.strokeState {
			g.strokeState = nonogram.CellEmpty
		}
	}
	if !g.pointerDown && !g.dragging {
		return
	}
	if cellX == g.lastCellX && cellY == g.lastCellY {
		return
	}

	next, corrected := g.correctedStrokeState(cellX, cellY, g.strokeState)
	if g.board.SetCell(cellX, cellY, next) {
		if corrected {
			g.timePenalty += 10 * time.Second
			g.penaltyFlashUntil = time.Now().Add(900 * time.Millisecond)
			g.correctFlashUntil = time.Now().Add(850 * time.Millisecond)
			g.correctFlashX = cellX
			g.correctFlashY = cellY
			playWebSFX("correct")
		} else if next == nonogram.CellFilled {
			playWebSFX("pencil")
		} else if next == nonogram.CellMarked || next == nonogram.CellEmpty {
			playWebSFX("eraser")
		}
		if nonogram.IsSolved(g.board, g.puzzle.Solution) {
			g.completePuzzle()
		}
	}
	g.dragging = true
	g.lastCellX = cellX
	g.lastCellY = cellY
}

func (g *Game) correctedStrokeState(cellX, cellY int, attempted nonogram.CellState) (nonogram.CellState, bool) {
	if !g.autoCorrect || attempted == nonogram.CellEmpty {
		return attempted, false
	}
	if attempted == nonogram.CellFilled && !g.puzzle.Solution[cellY][cellX] {
		return nonogram.CellMarked, true
	}
	if attempted == nonogram.CellMarked && g.puzzle.Solution[cellY][cellX] {
		return nonogram.CellFilled, true
	}
	return attempted, false
}

func (g *Game) updateRevealInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.retry()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.leaveReveal()
		return
	}

	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	if g.layout.retryButton.Contains(x, y) {
		g.retry()
		return
	}
	if g.layout.revealLevelsButton.Contains(x, y) {
		g.leaveReveal()
	}
}

func (g *Game) updateMainMenuInput() {
	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	switch {
	case mainLevelButton().Contains(x, y):
		g.mode = screenLevelSelect
	case mainCommunityButton().Contains(x, y):
		g.communityView = communityHome
		g.mode = screenCommunity
	case mainSettingsButton().Contains(x, y):
		g.mode = screenSettings
	}
}

func (g *Game) updateCommunityInput() {
	if raw := takeCommunityCloudDrafts(); raw != "" {
		if err := g.mergeCloudDrafts(raw); err != nil {
			g.showCommunityNotice("could not sync cloud drafts")
		}
	}
	if raw := takeCommunityCatalog(); raw != "" {
		if err := g.loadCommunityCatalog(raw); err != nil {
			g.showCommunityNotice("could not load community levels")
		}
	}
	if result := takeCommunityResult(); result != "" {
		g.showCommunityNotice(result)
	}
	if raw := takeCommunityImport(); raw != "" {
		if err := g.importCommunityPack(raw); err != nil {
			g.showCommunityNotice("import failed: " + err.Error())
		}
	}
	if g.communityView == communitySignIn && !communitySignedIn() {
		for _, char := range ebiten.AppendInputChars(nil) {
			if len(g.communityEmail) < 80 && (char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("@._+-", char)) {
				g.communityEmail += string(char)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.communityEmail) > 0 {
			g.communityEmail = g.communityEmail[:len(g.communityEmail)-1]
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.submitCommunitySignIn()
			return
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.communityBack()
		return
	}
	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	if communityBackButton().Contains(x, y) {
		g.communityBack()
		return
	}
	if g.communityView == communityHome && communityProfileBadgeButton().Contains(x, y) {
		if communitySignedIn() {
			g.openProfileEditor()
		} else {
			g.communityView = communitySignIn
		}
		return
	}
	if g.communityView == communityHome && communityAccountButton().Contains(x, y) {
		g.communityView = communitySignIn
		return
	}
	switch g.communityView {
	case communityHome:
		switch {
		case communityLevelsButton().Contains(x, y):
			g.communityView = communityBrowse
			g.communityPage = 0
			requestCommunityCatalog("levels")
		case communityPacksButton().Contains(x, y):
			g.communityView = communityPacks
		case communityCreateButton().Contains(x, y):
			g.communityView = communityCreate
		case communityMyArtButton().Contains(x, y):
			g.communityView = communityMyArt
			g.communityPage = 0
			g.syncLocalDrafts()
			requestCommunityCloudDrafts()
		}
	case communityCreate:
		switch {
		case communityNewButton().Contains(x, y):
			g.newCommunityDraft(10)
		case communityImportButton().Contains(x, y):
			if !requestCommunityImport() {
				g.showCommunityNotice("import is available in the web build")
			}
		}
	case communityMyArt:
		start := g.communityPage * communityDraftsPerPage
		for slot := 0; slot < communityDraftsPerPage; slot++ {
			if communityDraftEditButton(slot).Contains(x, y) {
				g.editCommunityDraft(start + slot)
				return
			}
			if communityDraftPublishButton(slot).Contains(x, y) {
				g.publishCommunityDraft(start + slot)
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityDraftsPerPage < len(g.communityLibrary.Drafts) {
			g.communityPage++
		}
	case communityBrowse:
		for slot := 0; slot < communityCatalogPerPage; slot++ {
			if communityCatalogPlayButton(slot).Contains(x, y) {
				g.playCommunityVersion(g.communityPage*communityCatalogPerPage + slot)
				return
			}
		}
	case communityPacks:
		if communityPackCreateButton().Contains(x, y) {
			g.openPackBuilder()
			return
		}
		for slot := 0; slot < 4; slot++ {
			if communityPackPlayButton(slot).Contains(x, y) {
				g.playLocalPack(slot)
				return
			}
			if communityPackPublishButton(slot).Contains(x, y) {
				g.publishLocalPack(slot)
				return
			}
		}
	case communityPackBuild:
		start := g.communityPage * communityPackDraftsPerPage
		for slot := 0; slot < communityPackDraftsPerPage; slot++ {
			if communityPackDraftButton(slot).Contains(x, y) {
				g.togglePackDraft(start + slot)
				return
			}
		}
		if communityPackDoneButton().Contains(x, y) {
			g.createLocalPack()
			return
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
			g.communityPage++
		}
	case communitySignIn:
		if communitySignedIn() {
			if communitySignOutButton().Contains(x, y) {
				requestCommunitySignOut()
				g.communityView = communityHome
			}
		} else if communitySendLinkButton().Contains(x, y) {
			g.submitCommunitySignIn()
		}
	}
}

func (g *Game) submitCommunitySignIn() {
	email := strings.TrimSpace(g.communityEmail)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		g.showCommunityNotice("enter a valid email address")
		return
	}
	if !requestCommunitySignIn(email) {
		g.showCommunityNotice("sign in is available in the web build")
	}
}

func (g *Game) communityBack() {
	if g.communityView == communityPackBuild {
		g.communityView = communityPacks
		return
	}
	if g.communityView == communityHome {
		g.mode = screenMainMenu
		return
	}
	g.communityView = communityHome
}

func (g *Game) updateEditorInput() {
	if title := takeEditorTitle(); title != "" {
		g.editor.Title = title
		_ = g.saveCurrentDraft(false)
	}
	if raw := takeEditorColorPicker(); raw != "" {
		if c, ok := parseEditorHexColor(raw); ok {
			g.editor.selectPaintColor(c)
			g.editor.Tool = editorToolPencil
		}
	}
	if raw := takeEditorImageImport(); raw != "" {
		g.pushEditorUndo()
		if err := g.editor.importPayload(raw); err != nil {
			g.showMenuNotice("import failed")
		} else {
			g.showMenuNotice("image imported")
		}
	}
	if raw := takeEditorPackImport(); raw != "" {
		g.pushEditorUndo()
		if editor, err := editorFromPackJSON(raw); err != nil {
			g.showMenuNotice("pack failed")
		} else {
			g.editor = editor
			g.showMenuNotice("pack imported")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.editorSizeOpen {
			g.editorSizeOpen = false
			return
		}
		if g.editingProfile {
			g.closeProfileEditor(false)
			return
		}
		_ = g.saveCurrentDraft(false)
		g.communityView = communityCreate
		g.mode = screenCommunity
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.undoEditor()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.editor.Tool = editorToolPencil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.editor.Tool = editorToolEraser
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.editor.Tool = editorToolFill
	}

	x, y, down, justPressed, justReleased := pointerState()
	if justReleased {
		g.editorPointer = false
		g.editorLastX = -1
		g.editorLastY = -1
	}
	if justPressed {
		if g.editorSizeOpen {
			if editorSizeButton().Contains(x, y) {
				g.editorSizeOpen = false
				return
			}
			for _, size := range editorSizes {
				if editorSizeOption(size).Contains(x, y) {
					g.resetEditor(size)
					g.editorSizeOpen = false
					return
				}
			}
			g.editorSizeOpen = false
			return
		}
		if g.handleEditorButton(x, y) {
			return
		}
		for i, c := range editorPalette {
			if editorPaletteRect(i).Contains(x, y) {
				g.editor.selectPaintColor(c)
				g.editor.Tool = editorToolPencil
				return
			}
		}
		cellX, cellY, ok := editorCellAt(g.editor, x, y)
		if ok {
			g.pushEditorUndo()
			g.editor.apply(cellX, cellY)
			g.editorPointer = true
			g.editorLastX = cellX
			g.editorLastY = cellY
			return
		}
	}
	if !down || !g.editorPointer || g.editor.Tool == editorToolFill || g.editor.Tool == editorToolEyedropper {
		return
	}
	cellX, cellY, ok := editorCellAt(g.editor, x, y)
	if !ok || (cellX == g.editorLastX && cellY == g.editorLastY) {
		return
	}
	g.editor.applyLine(g.editorLastX, g.editorLastY, cellX, cellY)
	g.editorLastX = cellX
	g.editorLastY = cellY
}

func (g *Game) handleEditorButton(x, y int) bool {
	if g.editingProfile {
		switch {
		case editorBackButton().Contains(x, y):
			g.closeProfileEditor(false)
		case profileSaveButton().Contains(x, y):
			g.closeProfileEditor(true)
		default:
			return false
		}
		return true
	}
	switch {
	case editorBackButton().Contains(x, y):
		_ = g.saveCurrentDraft(false)
		g.communityView = communityMyArt
		g.mode = screenCommunity
	case editorUndoButton().Contains(x, y):
		g.undoEditor()
	case editorSizeButton().Contains(x, y):
		g.editorSizeOpen = !g.editorSizeOpen
	case editorTitleButton().Contains(x, y):
		requestEditorTitle(g.editor.Title)
	case editorPreviewButton().Contains(x, y):
		g.loadEditorPuzzle()
	case editorSaveButton().Contains(x, y):
		g.saveEditor()
	case editorExportButton().Contains(x, y):
		g.exportEditor()
	case editorPencilButton().Contains(x, y):
		g.editor.Tool = editorToolPencil
	case editorEraserButton().Contains(x, y):
		g.editor.Tool = editorToolEraser
	case editorFillButton().Contains(x, y):
		g.editor.Tool = editorToolFill
	case editorEyeButton().Contains(x, y):
		g.editor.Tool = editorToolEyedropper
	case editorBeforeButton().Contains(x, y):
		g.editor.selectLayer(editorLayerBefore)
	case editorAfterButton().Contains(x, y):
		g.editor.selectLayer(editorLayerAfter)
	case editorLayerPreviewButton().Contains(x, y) && g.editor.Layer == editorLayerAfter:
		g.editorOnionSkin = !g.editorOnionSkin
	case editorRainbowRect().Contains(x, y):
		if g.editor.Layer == editorLayerBefore {
			g.editor.selectPaintColor(editorBeforeColor)
			break
		}
		if !requestEditorColorPicker(editorColorHex(g.editor.PaintColor)) {
			g.showMenuNotice("color picker unavailable")
		}
	default:
		return false
	}
	return true
}

func (g *Game) updateLevelSelectInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = screenMainMenu
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.prevLevelPage()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.nextLevelPage()
		return
	}

	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	if g.layout.levelBackButton.Contains(x, y) {
		g.mode = screenMainMenu
		return
	}
	if g.layout.levelPrevButton.Contains(x, y) {
		g.prevLevelPage()
		return
	}
	if g.layout.levelNextButton.Contains(x, y) {
		g.nextLevelPage()
		return
	}
	pageStart := g.levelPage * levelSelectPageSize
	for slot := 0; slot < levelSelectPageSize; slot++ {
		if levelTileRect(slot).Contains(x, y) {
			levelIndex := pageStart + slot
			if levelIndex < len(gameLevels) && gameLevels[levelIndex].Available {
				_ = g.loadLevel(levelIndex)
			} else {
				g.showMenuNotice("LW")
			}
			return
		}
	}
}

func (g *Game) updateSettingsInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = screenMainMenu
		return
	}

	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	switch {
	case g.layout.soundButton.Contains(x, y):
		g.audioEnabled = !g.audioEnabled
		setWebMusicMuted(!g.audioEnabled)
	case g.layout.autoCorrectButton.Contains(x, y):
		g.autoCorrect = !g.autoCorrect
	case g.layout.settingsCloseButton.Contains(x, y):
		g.mode = screenMainMenu
	}
}

func pointerState() (int, int, bool, bool, bool) {
	x, y := ebiten.CursorPosition()
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	justPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	justReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	touches := ebiten.AppendTouchIDs(nil)
	if len(touches) > 0 {
		tx, ty := ebiten.TouchPosition(touches[0])
		x, y = tx, ty
		down = true
		justPressed = inpututil.IsTouchJustReleased(touches[0]) == false && inpututil.TouchPressDuration(touches[0]) == 1
		justReleased = false
	}
	return x, y, down, justPressed, justReleased
}

type rect struct {
	x float64
	y float64
	w float64
	h float64
}

func (r rect) Contains(px, py int) bool {
	return float64(px) >= r.x && float64(px) <= r.x+r.w && float64(py) >= r.y && float64(py) <= r.y+r.h
}

func (r rect) ImageRect() image.Rectangle {
	return image.Rect(int(r.x), int(r.y), int(r.x+r.w), int(r.y+r.h))
}
