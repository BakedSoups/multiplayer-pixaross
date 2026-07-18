package game

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	communityDraftsPerPage     = 4
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
	case communityCreators:
		g.drawCommunityCreators(screen)
	case communityCreatorProfile:
		g.drawCommunityCreatorProfile(screen)
	case communityGalleryPack:
		g.drawCommunityGalleryPack(screen)
	case communityImportHelp:
		g.drawCommunityImportHelp(screen)
	default:
		g.drawCommunityHome(screen)
	}
	if time.Now().Before(g.communityNoticeUntil) {
		drawCenteredText(screen, g.communityNotice, rect{x: 36, y: 610, w: 468, h: 38}, colAccent)
	}
	drawButton(screen, communityBackButton(), "back")
}

func (g *Game) drawCommunityBrowse(screen *ebiten.Image) {
	drawCenteredText(screen, "GALLERY", rect{x: 100, y: 190, w: 340, h: 28}, colInk)
	drawButton(screen, communityGalleryArtButton(), "Art")
	drawButton(screen, communityGalleryPacksButton(), "Packs")
	drawButton(screen, communityGalleryNewButton(), "New")
	drawButton(screen, communityGalleryTopButton(), "Top")
	if g.galleryKind == "art" {
		drawRectOutline(screen, communityGalleryArtButton(), 2, colAccent)
	} else {
		drawRectOutline(screen, communityGalleryPacksButton(), 2, colAccent)
	}
	if g.gallerySort == "new" {
		drawRectOutline(screen, communityGalleryNewButton(), 2, colAccent)
	} else {
		drawRectOutline(screen, communityGalleryTopButton(), 2, colAccent)
	}
	if len(g.communityGallery) == 0 {
		drawCenteredText(screen, "No published work yet", rect{x: 60, y: 346, w: 420, h: 40}, colMuted)
		return
	}
	start := g.communityPage * communityCatalogPerPage
	for slot := 0; slot < communityCatalogPerPage && start+slot < len(g.communityGallery); slot++ {
		item := g.communityGallery[start+slot]
		r := communityGalleryCard(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		if item.Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		} else if len(item.Levels) > 0 && item.Levels[0].Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Levels[0].Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		}
		title := item.Title
		if len(title) > 15 {
			title = title[:15]
		}
		creator := item.CreatorName
		if len(creator) > 15 {
			creator = creator[:15]
		}
		drawText(screen, title, int(r.x+70), int(r.y+18), colInk)
		drawText(screen, creator, int(r.x+70), int(r.y+41), colMuted)
		drawButton(screen, communityGalleryOpenButton(slot), map[bool]string{true: "open", false: "play"}[item.Kind == "pack"])
		drawButton(screen, communityGalleryLikeButton(slot), fmt.Sprintf("+ %d", item.Likes))
		if item.Owned {
			drawButton(screen, communityGalleryPromoteButton(slot), "pin")
		}
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityCatalogPerPage < len(g.communityGallery) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityGalleryPack(screen *ebiten.Image) {
	if g.selectedGallery < 0 || g.selectedGallery >= len(g.communityGallery) {
		return
	}
	item := g.communityGallery[g.selectedGallery]
	drawCenteredText(screen, item.Title, rect{x: 70, y: 194, w: 400, h: 30}, colInk)
	drawCenteredText(screen, fmt.Sprintf("by %s   + %d", item.CreatorName, item.Likes), rect{x: 70, y: 224, w: 400, h: 24}, colMuted)
	for slot := 0; slot < 6 && slot < len(item.Levels); slot++ {
		level := item.Levels[slot]
		r := communityGalleryPackLevelButton(slot)
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		if level.Puzzle != nil {
			drawCommunityArtThumbnail(screen, level.Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		}
		title := level.Title
		if len(title) > 12 {
			title = title[:12]
		}
		drawCenteredText(screen, title, rect{x: r.x + 70, y: r.y + 8, w: r.w - 76, h: 28}, colInk)
		drawCenteredText(screen, "play", rect{x: r.x + 70, y: r.y + 38, w: r.w - 76, h: 26}, colAccent)
	}
}

func (g *Game) drawCommunityPacks(screen *ebiten.Image) {
	drawCenteredText(screen, "MY LIBRARY", rect{x: 100, y: 190, w: 340, h: 30}, colInk)
	drawLibraryTabs(screen, false)
	drawButton(screen, communityPackCreateButton(), "Create Pack")
	for i, pack := range g.communityLibrary.Packs {
		if i >= 4 {
			break
		}
		r := communityPackRect(i)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		title := pack.Title
		if len(title) > 10 {
			title = title[:10]
		}
		drawText(screen, title, int(r.x+10), int(r.y+20), colInk)
		packStatus := fmt.Sprintf("%d art", len(pack.Items))
		if string(pack.Status) == "published" {
			packStatus = "published"
		}
		if g.pendingPackPublishID == pack.ID {
			packStatus = "publishing"
		}
		drawText(screen, packStatus, int(r.x+10), int(r.y+44), colMuted)
		for art, item := range pack.Items {
			if art >= 20 {
				break
			}
			draft, ok := g.communityLibrary.Draft(item.LevelID)
			if !ok || draft.Puzzle == nil {
				continue
			}
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityPackArtPreview(i, art, len(pack.Items)))
		}
		drawButton(screen, communityPackPlayButton(i), "play")
		publishLabel := "publish"
		if g.pendingPackPublishID == pack.ID {
			publishLabel = "..."
		}
		drawButton(screen, communityPackPublishButton(i), publishLabel)
		drawButton(screen, communityPackDeleteButton(i), "x")
	}
}

func (g *Game) drawCommunityHome(screen *ebiten.Image) {
	drawButton(screen, communityLevelsButton(), "Community Gallery")
	drawButton(screen, communityMyArtButton(), "My Library")
	drawButton(screen, communityCreatorsButton(), "Creators")
}

func (g *Game) drawCommunityCreators(screen *ebiten.Image) {
	drawCenteredText(screen, "CREATORS", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
	if len(g.communityCreators) == 0 {
		drawCenteredText(screen, "No creators yet", rect{x: 80, y: 350, w: 380, h: 32}, colMuted)
		return
	}
	start := g.communityPage * 4
	for slot := 0; slot < 4 && start+slot < len(g.communityCreators); slot++ {
		creator := g.communityCreators[start+slot]
		r := communityCreatorButton(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		avatar := defaultCommunityProfilePixels
		if creator.AvatarPuzzle != nil {
			avatar = creator.AvatarPuzzle.RevealRaw
		}
		drawCommunityArtThumbnail(screen, avatar, rect{x: r.x + 8, y: r.y + 8, w: 56, h: 56})
		name := creator.DisplayName
		if len(name) > 20 {
			name = name[:20]
		}
		drawText(screen, name, int(r.x+76), int(r.y+25), colInk)
		drawText(screen, fmt.Sprintf("%d art", len(creator.Levels)), int(r.x+76), int(r.y+50), colMuted)
		for art := 0; art < 3 && art < len(creator.Levels); art++ {
			if creator.Levels[art].Puzzle != nil {
				drawCommunityArtThumbnail(screen, creator.Levels[art].Puzzle.RevealRaw, communityCreatorPreviewRect(slot, art))
			}
		}
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*4 < len(g.communityCreators) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityCreatorProfile(screen *ebiten.Image) {
	if g.selectedCreator < 0 || g.selectedCreator >= len(g.communityCreators) {
		drawCenteredText(screen, "CREATOR", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
		return
	}
	creator := g.communityCreators[g.selectedCreator]
	avatar := defaultCommunityProfilePixels
	if creator.AvatarPuzzle != nil {
		avatar = creator.AvatarPuzzle.RevealRaw
	}
	drawCommunityArtThumbnail(screen, avatar, rect{x: 54, y: 206, w: 82, h: 82})
	drawText(screen, creator.DisplayName, 154, 242, colInk)
	drawText(screen, fmt.Sprintf("%d published", len(creator.Levels)), 154, 269, colMuted)
	contentY := float64(310)
	if len(creator.Featured) > 0 {
		featured := creator.Featured[0]
		drawText(screen, "PROMOTED", 48, 310, colAccent)
		r := communityCreatorFeaturedButton()
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colAccent)
		if featured.Puzzle != nil {
			drawCommunityArtThumbnail(screen, featured.Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		} else if len(featured.Levels) > 0 && featured.Levels[0].Puzzle != nil {
			drawCommunityArtThumbnail(screen, featured.Levels[0].Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		}
		drawText(screen, featured.Title, int(r.x+76), int(r.y+23), colInk)
		drawText(screen, fmt.Sprintf("%s  + %d", featured.Kind, featured.Likes), int(r.x+76), int(r.y+48), colMuted)
		contentY = 410
	}
	start := g.communityPage * 4
	for slot := 0; slot < 4 && start+slot < len(creator.Levels); slot++ {
		level := creator.Levels[start+slot]
		r := communityCreatorLevelButtonAt(slot, contentY)
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		if level.Puzzle != nil {
			drawCommunityArtThumbnail(screen, level.Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		}
		title := level.Title
		if len(title) > 12 {
			title = title[:12]
		}
		drawCenteredText(screen, title, rect{x: r.x + 70, y: r.y + 8, w: r.w - 76, h: 28}, colInk)
		drawCenteredText(screen, "play", rect{x: r.x + 70, y: r.y + 38, w: r.w - 76, h: 26}, colAccent)
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*4 < len(creator.Levels) {
		drawButton(screen, communityNextButton(), "next")
	}
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
	drawButton(screen, communityCreatePackButton(), "Create Pack")
	drawButton(screen, communityImportHelpButton(), "Import Instructions")
}

func (g *Game) drawCommunityImportHelp(screen *ebiten.Image) {
	drawCenteredText(screen, "SPRITE SHEET IMPORT", rect{x: 48, y: 190, w: 444, h: 34}, colInk)
	panel := rect{x: 48, y: 238, w: 444, h: 342}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 2, colGridHeavy)
	drawText(screen, "ANY DRAWING APP", 70, 270, colAccent)
	drawText(screen, "1. Use 8, 10, 15, or 20 px squares", 70, 302, colInk)
	drawText(screen, "2. Pair Before/After left-right or up-down", 70, 334, colInk)
	drawText(screen, "3. Export one PNG sprite sheet", 70, 366, colInk)
	drawText(screen, "Before is black. After keeps color.", 70, 404, colMuted)
	drawText(screen, "ASEPRITE MULTI-ART", 70, 450, colAccent)
	drawText(screen, "Export PNG + JSON (array data)", 70, 482, colInk)
	drawText(screen, "Name: flower_before / flower_after", 70, 514, colInk)
	drawText(screen, "Each named pair becomes one artwork.", 70, 546, colMuted)
}

func (g *Game) drawCommunitySignIn(screen *ebiten.Image) {
	drawCenteredText(screen, "ACCOUNT", rect{x: 100, y: 218, w: 340, h: 32}, colInk)
	panel := rect{x: 72, y: 254, w: 396, h: 328}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 3, colGridHeavy)
	if communitySignedIn() {
		drawCommunityArtThumbnail(screen, g.profileArt.rawPixels(editorLayerAfter), rect{x: 218, y: 304, w: 104, h: 104})
		drawCenteredText(screen, "Signed in", rect{x: 120, y: 424, w: 300, h: 30}, colInk)
		drawButton(screen, communitySignOutButton(), "sign out")
		return
	}
	drawCommunityArtThumbnail(screen, defaultCommunityProfilePixels, rect{x: 230, y: 270, w: 80, h: 80})
	drawButton(screen, communityGoogleButton(), "continue with Google")
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
	drawCenteredText(screen, "MY LIBRARY", rect{x: 100, y: 190, w: 340, h: 30}, colInk)
	drawLibraryTabs(screen, true)
	drawButton(screen, communityArtCreateButton(), "Create or Import")
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
		status := string(draft.Status)
		if status == "" {
			status = "draft"
		}
		statusText := fmt.Sprintf("%dx%d %s", draft.Puzzle.Width, draft.Puzzle.Height, status)
		if g.pendingPublishID == draft.ID {
			statusText = "publishing"
		}
		drawText(screen, statusText, int(r.x+116), int(r.y+43), colMuted)
		drawButton(screen, communityDraftEditButton(slot), "edit")
		publishLabel := "pub"
		if g.pendingPublishID == draft.ID {
			publishLabel = "..."
		}
		drawButton(screen, communityDraftPublishButton(slot), publishLabel)
		drawButton(screen, communityDraftDeleteButton(slot), "x")
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityDraftsPerPage < len(g.communityLibrary.Drafts) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func drawLibraryTabs(screen *ebiten.Image, art bool) {
	drawButton(screen, communityLibraryArtTab(), "Art")
	drawButton(screen, communityLibraryPacksTab(), "Packs")
	if art {
		drawRectOutline(screen, communityLibraryArtTab(), 2, colAccent)
	} else {
		drawRectOutline(screen, communityLibraryPacksTab(), 2, colAccent)
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
func communityLevelsButton() rect       { return rect{x: 108, y: 314, w: 324, h: 50} }
func communityPacksButton() rect        { return rect{x: 128, y: 348, w: 284, h: 46} }
func communityCreateButton() rect       { return rect{x: 128, y: 410, w: 284, h: 46} }
func communityMyArtButton() rect        { return rect{x: 108, y: 388, w: 324, h: 50} }
func communityCreatorsButton() rect     { return rect{x: 108, y: 462, w: 324, h: 50} }
func communityNewButton() rect          { return rect{x: 104, y: 270, w: 332, h: 48} }
func communityImportButton() rect       { return rect{x: 104, y: 338, w: 332, h: 48} }
func communityCreatePackButton() rect   { return rect{x: 104, y: 406, w: 332, h: 48} }
func communityImportHelpButton() rect   { return rect{x: 104, y: 474, w: 332, h: 48} }
func communityLibraryArtTab() rect      { return rect{x: 126, y: 226, w: 142, h: 36} }
func communityLibraryPacksTab() rect    { return rect{x: 272, y: 226, w: 142, h: 36} }
func communityArtCreateButton() rect    { return rect{x: 154, y: 270, w: 232, h: 38} }
func communityDraftRect(slot int) rect {
	return rect{x: 54, y: 234 + float64(slot)*88, w: 432, h: 74}
}
func communityMyArtRect(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 44 + float64(column)*234, y: 320 + float64(row)*136, w: 218, h: 120}
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
	return rect{x: r.x + 112, y: r.y + 62, w: 34, h: 32}
}
func communityDraftPublishButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 150, y: r.y + 62, w: 38, h: 32}
}
func communityDraftDeleteButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 192, y: r.y + 62, w: 20, h: 32}
}
func communityPrevButton() rect { return rect{x: 62, y: 604, w: 92, h: 38} }
func communityNextButton() rect { return rect{x: 386, y: 604, w: 92, h: 38} }
func communityCatalogPlayButton(slot int) rect {
	r := communityDraftRect(slot)
	return rect{x: r.x + r.w - 90, y: r.y + 17, w: 72, h: 40}
}
func communityGalleryArtButton() rect   { return rect{x: 62, y: 226, w: 94, h: 34} }
func communityGalleryPacksButton() rect { return rect{x: 160, y: 226, w: 94, h: 34} }
func communityGalleryNewButton() rect   { return rect{x: 286, y: 226, w: 92, h: 34} }
func communityGalleryTopButton() rect   { return rect{x: 382, y: 226, w: 92, h: 34} }
func communityGalleryCard(slot int) rect {
	return rect{x: 48, y: 270 + float64(slot)*78, w: 444, h: 70}
}
func communityGalleryOpenButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 272, y: r.y + 37, w: 56, h: 27}
}
func communityGalleryLikeButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 334, y: r.y + 37, w: 48, h: 27}
}
func communityGalleryPromoteButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 388, y: r.y + 37, w: 48, h: 27}
}
func communityGalleryPackLevelButton(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 46 + float64(column)*232, y: 270 + float64(row)*92, w: 216, h: 78}
}
func communityCreatorFeaturedButton() rect { return rect{x: 46, y: 334, w: 448, h: 70} }
func communityCreatorLevelButtonAt(slot int, y float64) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 46 + float64(column)*232, y: y + float64(row)*92, w: 216, h: 78}
}
func communityPackCreateButton() rect { return rect{x: 154, y: 270, w: 232, h: 38} }
func communityPackRect(slot int) rect { return rect{x: 44, y: 318 + float64(slot)*70, w: 452, h: 64} }
func communityPackArtPreview(slot, art, count int) rect {
	r := communityPackRect(slot)
	if count <= 5 {
		return rect{x: r.x + 104 + float64(art)*31, y: r.y + 17, w: 28, h: 28}
	}
	column := art % 10
	row := art / 10
	return rect{x: r.x + 104 + float64(column)*17, y: r.y + 9 + float64(row)*25, w: 16, h: 16}
}
func communityPackPlayButton(slot int) rect {
	r := communityPackRect(slot)
	return rect{x: r.x + 280, y: r.y + 14, w: 54, h: 36}
}
func communityPackPublishButton(slot int) rect {
	r := communityPackRect(slot)
	return rect{x: r.x + 340, y: r.y + 14, w: 68, h: 36}
}
func communityPackDeleteButton(slot int) rect {
	r := communityPackRect(slot)
	return rect{x: r.x + 414, y: r.y + 14, w: 28, h: 36}
}
func communityGoogleButton() rect   { return rect{x: 122, y: 370, w: 296, h: 42} }
func communityEmailInput() rect     { return rect{x: 102, y: 434, w: 336, h: 44} }
func communitySendLinkButton() rect { return rect{x: 142, y: 500, w: 256, h: 42} }
func communitySignOutButton() rect  { return rect{x: 174, y: 468, w: 192, h: 42} }
func communityPackDraftButton(slot int) rect {
	return rect{x: 66, y: 246 + float64(slot)*72, w: 408, h: 64}
}
func communityPackDoneButton() rect { return rect{x: 160, y: 552, w: 220, h: 42} }
func communityCreatorButton(slot int) rect {
	return rect{x: 48, y: 234 + float64(slot)*88, w: 444, h: 76}
}
func communityCreatorPreviewRect(slot, art int) rect {
	r := communityCreatorButton(slot)
	return rect{x: r.x + 292 + float64(art)*48, y: r.y + 16, w: 42, h: 42}
}
func communityCreatorLevelButton(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 46 + float64(column)*232, y: 310 + float64(row)*92, w: 216, h: 78}
}

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
