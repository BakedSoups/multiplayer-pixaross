package game

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	communityDraftsPerPage     = 6
	communityCatalogPerPage    = 4
	communityPackDraftsPerPage = 4
)

func (g *Game) drawCommunity(screen *ebiten.Image) {
	drawMenuBackdrop(screen)
	drawScaledTextCentered(screen, "COMMUNITY", rect{x: 76, y: 42, w: 388, h: 52}, 2.1, colInk)
	if g.communityView == communityHome {
		g.drawCommunityAccount(screen)
	}
	switch g.communityView {
	case communityBrowse:
		g.drawCommunityBrowse(screen)
	case communityPacks:
		g.drawCommunityPacks(screen)
	case communityCreate:
		g.drawCommunityCreate(screen)
	case communityMyArt:
		g.drawCommunityMyArt(screen)
	case communitySignIn:
		g.drawCommunitySignIn(screen)
	case communityPackBuild:
		g.drawCommunityPackBuilder(screen)
	default:
		g.drawCommunityHome(screen)
	}
	if time.Now().Before(g.communityNoticeUntil) {
		drawCenteredText(screen, g.communityNotice, rect{x: 36, y: 610, w: 468, h: 38}, colAccent)
	}
	drawButton(screen, communityBackButton(), "back")
}

func (g *Game) drawCommunityBrowse(screen *ebiten.Image) {
	drawCenteredText(screen, "LEVELS", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
	if len(g.communityCatalog) == 0 {
		drawCenteredText(screen, communityFetchStatus(), rect{x: 60, y: 326, w: 420, h: 70}, colMuted)
		return
	}
	start := g.communityPage * communityCatalogPerPage
	for slot := 0; slot < communityCatalogPerPage && start+slot < len(g.communityCatalog); slot++ {
		version := g.communityCatalog[start+slot]
		r := communityDraftRect(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawText(screen, version.Title, int(r.x+14), int(r.y+20), colInk)
		if version.Puzzle != nil {
			drawText(screen, fmt.Sprintf("%dx%d", version.Puzzle.Width, version.Puzzle.Height), int(r.x+14), int(r.y+45), colMuted)
		}
		drawButton(screen, communityCatalogPlayButton(slot), "play")
	}
}

func (g *Game) drawCommunityPacks(screen *ebiten.Image) {
	drawCenteredText(screen, "PACKS", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
	drawButton(screen, communityPackCreateButton(), "Create from My Art")
	for i, pack := range g.communityLibrary.Packs {
		if i >= 4 {
			break
		}
		r := rect{x: 70, y: 302 + float64(i)*66, w: 400, h: 52}
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawText(screen, pack.Title, int(r.x+14), int(r.y+18), colInk)
		drawText(screen, fmt.Sprintf("%d/%d", len(pack.Progress.CompletedLevelIDs), len(pack.Items)), int(r.x+220), int(r.y+18), colMuted)
		drawButton(screen, communityPackPlayButton(i), "play")
		drawButton(screen, communityPackPublishButton(i), "publish")
	}
}

func (g *Game) drawCommunityHome(screen *ebiten.Image) {
	drawButton(screen, communityLevelsButton(), "Levels")
	drawButton(screen, communityPacksButton(), "Packs")
	drawButton(screen, communityCreateButton(), "Create")
	drawButton(screen, communityMyArtButton(), fmt.Sprintf("My Art  %d", len(g.communityLibrary.Drafts)))
}

func (g *Game) drawCommunityAccount(screen *ebiten.Image) {
	drawButton(screen, communityAccountButton(), communityAccountLabel())
	pixels := defaultCommunityProfilePixels
	if communitySignedIn() {
		pixels = g.profileArt.rawPixels(editorLayerAfter)
	}
	drawCommunityArtThumbnail(screen, pixels, communityProfileBadgeButton())
}

func (g *Game) drawCommunityCreate(screen *ebiten.Image) {
	drawCenteredText(screen, "CREATE", rect{x: 100, y: 202, w: 340, h: 34}, colInk)
	drawButton(screen, communityNewButton(), "New Drawing")
	drawButton(screen, communityImportButton(), "Import Sprite Sheet")
	drawCenteredText(screen, "PNG sheets or Aseprite PNG + JSON", rect{x: 70, y: 420, w: 400, h: 34}, colMuted)
}

func (g *Game) drawCommunitySignIn(screen *ebiten.Image) {
	drawCenteredText(screen, "ACCOUNT", rect{x: 100, y: 218, w: 340, h: 32}, colInk)
	panel := rect{x: 72, y: 272, w: 396, h: 256}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 3, colGridHeavy)
	if communitySignedIn() {
		drawCommunityArtThumbnail(screen, g.profileArt.rawPixels(editorLayerAfter), rect{x: 218, y: 304, w: 104, h: 104})
		drawCenteredText(screen, "Signed in", rect{x: 120, y: 424, w: 300, h: 30}, colInk)
		drawButton(screen, communitySignOutButton(), "sign out")
		return
	}
	drawCommunityArtThumbnail(screen, defaultCommunityProfilePixels, rect{x: 222, y: 294, w: 96, h: 96})
	input := communityEmailInput()
	drawRounded(screen, input, 4, colPanel)
	drawRectOutline(screen, input, 3, colGridHeavy)
	email := g.communityEmail
	cursorLength := len(email)
	if email == "" {
		email = "email address"
		drawText(screen, email, int(input.x+12), int(input.y+26), colMuted)
	} else {
		if len(email) > 34 {
			email = email[len(email)-34:]
		}
		drawText(screen, email, int(input.x+12), int(input.y+26), colInk)
	}
	if time.Now().UnixMilli()/500%2 == 0 {
		cursorX := input.x + 12 + float64(cursorLength)*8
		vector.DrawFilledRect(screen, float32(cursorX), float32(input.y+12), 2, 20, colAccent, false)
	}
	drawButton(screen, communitySendLinkButton(), "email sign-in link")
}

func (g *Game) drawCommunityPackBuilder(screen *ebiten.Image) {
	drawCenteredText(screen, fmt.Sprintf("SELECT ART  %d/20", len(g.packSelection)), rect{x: 80, y: 204, w: 380, h: 32}, colInk)
	start := g.communityPage * communityPackDraftsPerPage
	if start >= len(g.communityLibrary.Drafts) {
		drawCenteredText(screen, "Create some art first", rect{x: 80, y: 360, w: 380, h: 30}, colMuted)
	} else {
		for slot := 0; slot < communityPackDraftsPerPage && start+slot < len(g.communityLibrary.Drafts); slot++ {
			draft := g.communityLibrary.Drafts[start+slot]
			r := communityPackDraftButton(slot)
			drawRounded(screen, r, 5, colWhite)
			drawRectOutline(screen, r, 2, colGridHeavy)
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 7, w: 50, h: 50})
			title := draft.Title
			if len(title) > 24 {
				title = title[:24]
			}
			drawText(screen, title, int(r.x+72), int(r.y+26), colInk)
			box := rect{x: r.x + r.w - 48, y: r.y + 13, w: 36, h: 36}
			drawRounded(screen, box, 3, colPanel)
			drawRectOutline(screen, box, 2, colGridHeavy)
			if g.packSelection[draft.ID] {
				drawCenteredText(screen, "x", box, colAccent)
			}
		}
	}
	drawButton(screen, communityPackDoneButton(), fmt.Sprintf("create pack  %d", len(g.packSelection)))
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityMyArt(screen *ebiten.Image) {
	drawCenteredText(screen, "MY ART", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
	start := g.communityPage * communityDraftsPerPage
	if start >= len(g.communityLibrary.Drafts) {
		drawCenteredText(screen, "No drafts yet", rect{x: 80, y: 350, w: 380, h: 32}, colMuted)
		return
	}
	for slot := 0; slot < communityDraftsPerPage && start+slot < len(g.communityLibrary.Drafts); slot++ {
		draft := g.communityLibrary.Drafts[start+slot]
		r := communityMyArtRect(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawCommunityArtThumbnail(screen, draft.Puzzle.SkeletonRaw, communityMyArtBeforePreviewRect(slot))
		drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityMyArtAfterPreviewRect(slot))
		title := draft.Title
		if len(title) > 24 {
			title = title[:24]
		}
		drawText(screen, title, int(r.x+8), int(r.y+16), colInk)
		status := "draft"
		if draft.Playtested {
			status = "tested"
		}
		drawText(screen, fmt.Sprintf("%dx%d %s", draft.Puzzle.Width, draft.Puzzle.Height, status), int(r.x+116), int(r.y+43), colMuted)
		drawButton(screen, communityDraftEditButton(slot), "edit")
		drawButton(screen, communityDraftPublishButton(slot), "publish")
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityDraftsPerPage < len(g.communityLibrary.Drafts) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func drawCommunityArtThumbnail(screen *ebiten.Image, pixels [][]string, frame rect) {
	drawRounded(screen, frame, 4, colPanel)
	drawRectOutline(screen, frame, 2, colGridHeavy)
	if len(pixels) == 0 {
		return
	}
	height := len(pixels)
	width := len(pixels[0])
	if width == 0 {
		return
	}
	cell := minFloat((frame.w-6)/float64(width), (frame.h-6)/float64(height))
	originX := frame.x + (frame.w-cell*float64(width))/2
	originY := frame.y + (frame.h-cell*float64(height))/2
	for y, row := range pixels {
		for x, value := range row {
			c, ok := parseEditorHexColor(value)
			if !ok || c.A == 0 {
				continue
			}
			vector.DrawFilledRect(screen, float32(originX+float64(x)*cell), float32(originY+float64(y)*cell), float32(cell+0.25), float32(cell+0.25), c, false)
		}
	}
}

func (g *Game) drawCommunityEmpty(screen *ebiten.Image, title, message string) {
	drawCenteredText(screen, title, rect{x: 100, y: 212, w: 340, h: 34}, colInk)
	drawCenteredText(screen, message, rect{x: 60, y: 326, w: 420, h: 70}, colMuted)
}

func communityBackButton() rect         { return rect{x: 202, y: 674, w: 136, h: 42} }
func communityAccountButton() rect      { return rect{x: 322, y: 208, w: 98, h: 40} }
func communityProfileBadgeButton() rect { return rect{x: 430, y: 198, w: 58, h: 58} }
func communityLevelsButton() rect       { return rect{x: 128, y: 286, w: 284, h: 46} }
func communityPacksButton() rect        { return rect{x: 128, y: 348, w: 284, h: 46} }
func communityCreateButton() rect       { return rect{x: 128, y: 410, w: 284, h: 46} }
func communityMyArtButton() rect        { return rect{x: 128, y: 472, w: 284, h: 46} }
func communityNewButton() rect          { return rect{x: 104, y: 270, w: 332, h: 48} }
func communityImportButton() rect       { return rect{x: 104, y: 338, w: 332, h: 48} }
func communityDraftRect(slot int) rect {
	return rect{x: 54, y: 234 + float64(slot)*88, w: 432, h: 74}
}
func communityMyArtRect(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 44 + float64(column)*234, y: 228 + float64(row)*120, w: 218, h: 108}
}
func communityMyArtBeforePreviewRect(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 8, y: r.y + 30, w: 48, h: 48}
}
func communityMyArtAfterPreviewRect(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 62, y: r.y + 30, w: 48, h: 48}
}
func communityDraftEditButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 116, y: r.y + 62, w: 38, h: 32}
}
func communityDraftPublishButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 158, y: r.y + 62, w: 52, h: 32}
}
func communityPrevButton() rect { return rect{x: 62, y: 604, w: 92, h: 38} }
func communityNextButton() rect { return rect{x: 386, y: 604, w: 92, h: 38} }
func communityCatalogPlayButton(slot int) rect {
	r := communityDraftRect(slot)
	return rect{x: r.x + r.w - 90, y: r.y + 17, w: 72, h: 40}
}
func communityPackCreateButton() rect { return rect{x: 92, y: 240, w: 356, h: 44} }
func communityPackPlayButton(slot int) rect {
	return rect{x: 316, y: 309 + float64(slot)*66, w: 62, h: 38}
}
func communityPackPublishButton(slot int) rect {
	return rect{x: 384, y: 309 + float64(slot)*66, w: 74, h: 38}
}
func communityEmailInput() rect     { return rect{x: 102, y: 406, w: 336, h: 44} }
func communitySendLinkButton() rect { return rect{x: 142, y: 468, w: 256, h: 42} }
func communitySignOutButton() rect  { return rect{x: 174, y: 468, w: 192, h: 42} }
func communityPackDraftButton(slot int) rect {
	return rect{x: 66, y: 246 + float64(slot)*72, w: 408, h: 64}
}
func communityPackDoneButton() rect { return rect{x: 160, y: 552, w: 220, h: 42} }

var defaultCommunityProfilePixels = func() [][]string {
	rows := []string{
		"0000110000",
		"0001111000",
		"0001111000",
		"0000110000",
		"0011111100",
		"0111111110",
		"0111111110",
		"0100000010",
		"0100000010",
		"0000000000",
	}
	pixels := make([][]string, len(rows))
	for y, row := range rows {
		pixels[y] = make([]string, len(row))
		for x, value := range row {
			if value == '1' {
				pixels[y][x] = "#000000FF"
			}
		}
	}
	return pixels
}()
