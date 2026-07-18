package game

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	communityDraftsPerPage     = 3
	communityCatalogPerPage    = 4
	communityPackDraftsPerPage = 4
)

func (g *Game) drawCommunity(screen *ebiten.Image) {
	drawCommunityBackdrop(screen)
	drawScaledTextCentered(screen, "GLOBAL COMMUNITY", rect{x: 76, y: 42, w: 388, h: 52}, 1.65, colInk)
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
	case communityPublished:
		g.drawCommunityPublished(screen)
	case communityPublishSetup:
		g.drawCommunityPublishSetup(screen)
	case communityImportPreview:
		g.drawCommunityImportPreview(screen)
	case communityNewArtSetup:
		g.drawCommunityNewArtSetup(screen)
	case communityPackSetup:
		g.drawCommunityPackSetup(screen)
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
	drawSelectedButton(screen, communityGalleryAllButton(), "All", g.galleryKind == "all")
	drawSelectedButton(screen, communityGalleryArtButton(), "Art", g.galleryKind == "art")
	drawSelectedButton(screen, communityGalleryPacksButton(), "Packs", g.galleryKind == "pack")
	drawSelectedButton(screen, communityGalleryNewButton(), "New", g.gallerySort == "new")
	drawSelectedButton(screen, communityGalleryTopButton(), "Top", g.gallerySort == "top")
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
		if len(item.PreviewRaw) > 0 {
			drawCommunityArtThumbnail(screen, item.PreviewRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		} else if item.Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		} else if len(item.Levels) > 0 && item.Levels[0].Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Levels[0].Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		}
		creator := item.CreatorName
		if len(creator) > 15 {
			creator = creator[:15]
		}
		avatar := defaultCommunityProfilePixels
		if item.AvatarPuzzle != nil {
			avatar = item.AvatarPuzzle.RevealRaw
		}
		avatarRect := rect{x: r.x + r.w - 42, y: r.y + 5, w: 32, h: 26}
		drawText(screen, creator, int(avatarRect.x-8-float64(len(creator)*8)), int(r.y+22), colMuted)
		drawCommunityArtThumbnail(screen, avatar, avatarRect)
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

func drawSelectedButton(screen *ebiten.Image, r rect, label string, selected bool) {
	if selected {
		drawRounded(screen, inset(r, -3), 9, colAccent)
	}
	drawButton(screen, r, label)
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
	if creator.Bio != "" {
		bio := creator.Bio
		if len(bio) > 42 {
			bio = bio[:42]
		}
		drawText(screen, bio, 154, 292, colMuted)
	}
	contentY := float64(310)
	if len(creator.Featured) > 0 {
		featured := creator.Featured[0]
		drawPixelFavoriteStar(screen, 48, 304)
		drawText(screen, "FAVORITE PICROSS", 72, 310, colAccent)
		r := communityCreatorFeaturedButton()
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colAccent)
		if featured.Puzzle != nil {
			drawCommunityArtThumbnail(screen, featured.Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		} else if len(featured.Levels) > 0 && featured.Levels[0].Puzzle != nil {
			drawCommunityArtThumbnail(screen, featured.Levels[0].Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		}
		drawText(screen, featured.Title, int(r.x+76), int(r.y+23), colInk)
		likeLabel := "LIKES"
		if featured.Likes == 1 {
			likeLabel = "LIKE"
		}
		drawText(screen, fmt.Sprintf("%s   %d %s", strings.ToUpper(featured.Kind), featured.Likes, likeLabel), int(r.x+76), int(r.y+48), colMuted)
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

func drawPixelFavoriteStar(screen *ebiten.Image, x, y int) {
	vector.DrawFilledRect(screen, float32(x+8), float32(y), 6, 22, colAccent, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+8), 22, 6, colAccent, false)
	vector.DrawFilledRect(screen, float32(x+4), float32(y+4), 14, 14, colAccent, false)
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
	drawButton(screen, communityNewButton(), "Create New Art")
	drawButton(screen, communityImportButton(), "Import Sprite Sheet")
	drawButton(screen, communityImportHelpButton(), "?")
}

func (g *Game) drawCommunityImportHelp(screen *ebiten.Image) {
	drawCenteredText(screen, "SPRITE SHEET IMPORT", rect{x: 48, y: 190, w: 444, h: 34}, colInk)
	panel := rect{x: 48, y: 230, w: 444, h: 370}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 2, colGridHeavy)
	drawCenteredText(screen, "BEFORE", rect{x: 76, y: 246, w: 112, h: 24}, colMuted)
	drawCenteredText(screen, "AFTER", rect{x: 198, y: 246, w: 112, h: 24}, colMuted)
	drawCommunityArtThumbnail(screen, importExampleBefore, rect{x: 88, y: 272, w: 88, h: 88})
	drawCommunityArtThumbnail(screen, importExampleAfter, rect{x: 210, y: 272, w: 88, h: 88})
	drawText(screen, "black shape", 316, 300, colInk)
	drawText(screen, "color reveal", 316, 332, colInk)
	drawText(screen, "PNG SHEETS", 70, 392, colAccent)
	drawText(screen, "8, 10, 15, or 20 px square frames", 70, 422, colInk)
	drawText(screen, "Pair Before/After left-right or up-down", 70, 450, colInk)
	drawText(screen, "ASEPRITE", 70, 492, colAccent)
	drawText(screen, "Export PNG + JSON (array data)", 70, 522, colInk)
	drawText(screen, "flower_before / flower_after", 70, 550, colInk)
	drawText(screen, "Each pair previews before import.", 70, 578, colMuted)
}

func (g *Game) drawCommunityImportPreview(screen *ebiten.Image) {
	drawCenteredText(screen, "IMPORT PREVIEW", rect{x: 70, y: 190, w: 400, h: 30}, colInk)
	for slot, puzzle := range g.communityImportPack.Levels {
		if slot >= 6 || puzzle == nil {
			break
		}
		r := communityImportPreviewRect(slot)
		drawCommunityArtThumbnail(screen, puzzle.RevealRaw, r)
	}
	drawCenteredText(screen, fmt.Sprintf("%d artwork", len(g.communityImportPack.Levels)), rect{x: 90, y: 494, w: 360, h: 30}, colMuted)
	drawButton(screen, communityImportConfirmButton(), "Import to My Art")
}

func (g *Game) drawCommunitySignIn(screen *ebiten.Image) {
	drawCenteredText(screen, "ACCOUNT", rect{x: 100, y: 218, w: 340, h: 32}, colInk)
	panel := rect{x: 72, y: 254, w: 396, h: 380}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 3, colGridHeavy)
	if communitySignedIn() {
		drawCommunityArtThumbnail(screen, g.profileArt.rawPixels(editorLayerAfter), rect{x: 226, y: 278, w: 88, h: 88})
		drawCenteredText(screen, "Signed in", rect{x: 120, y: 374, w: 300, h: 30}, colInk)
		drawPublishField(screen, communityAccountBioField(), "Bio", g.profileBioDraft, g.profileBioEditing)
		drawButton(screen, communityAccountBioSaveButton(), "save bio")
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
	drawPublishField(screen, communityArtSearchField(), "Search", g.artSearch, g.artSearchActive)
	drawButton(screen, communityArtCreateButton(), "Create")
	indexes := g.filteredCommunityDraftIndexes()
	start := g.communityPage * communityDraftsPerPage
	if start >= len(indexes) {
		drawCenteredText(screen, "No matching art", rect{x: 80, y: 350, w: 380, h: 32}, colMuted)
		return
	}
	for slot := 0; slot < communityDraftsPerPage && start+slot < len(indexes); slot++ {
		draft := g.communityLibrary.Drafts[indexes[start+slot]]
		r := communityMyArtRect(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawCommunityArtThumbnail(screen, draft.Puzzle.SkeletonRaw, communityMyArtBeforePreviewRect(slot))
		drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityMyArtAfterPreviewRect(slot))
		title := draft.Title
		if len(title) > 16 {
			title = title[:16]
		}
		drawText(screen, title, int(r.x+158), int(r.y+27), colInk)
		status := string(draft.Status)
		if status == "" {
			status = "draft"
		}
		statusText := fmt.Sprintf("%dx%d %s", draft.Puzzle.Width, draft.Puzzle.Height, status)
		if g.pendingPublishID == draft.ID {
			statusText = "publishing"
		}
		drawText(screen, statusText, int(r.x+158), int(r.y+54), colMuted)
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
	if (g.communityPage+1)*communityDraftsPerPage < len(indexes) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityNewArtSetup(screen *ebiten.Image) {
	drawCenteredText(screen, "TITLE YOUR ART", rect{x: 80, y: 214, w: 380, h: 34}, colInk)
	drawPublishField(screen, communityNewArtTitleField(), "Title", g.newArtTitle, true)
	drawButton(screen, communityNewArtStartButton(), "Start Drawing")
}

func (g *Game) drawCommunityPackSetup(screen *ebiten.Image) {
	drawCenteredText(screen, "PACK DETAILS", rect{x: 80, y: 190, w: 380, h: 30}, colInk)
	drawCenteredText(screen, "Choose the artwork shown as your pack cover", rect{x: 60, y: 216, w: 420, h: 20}, colMuted)
	for art, item := range g.packSetupItems {
		if art >= 8 {
			break
		}
		if draft, ok := g.communityLibrary.Draft(item.LevelID); ok && draft.Puzzle != nil {
			if art == g.packSetupPreview {
				drawRounded(screen, inset(communityPackSetupPreview(art), -4), 7, colAccent)
			}
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityPackSetupPreview(art))
		}
	}
	if len(g.packSetupPreviewRaw) > 0 {
		drawCommunityArtThumbnail(screen, g.packSetupPreviewRaw, rect{x: 242, y: 302, w: 56, h: 56})
	}
	drawButton(screen, communityPackUploadCoverButton(), "or upload cover")
	drawPublishField(screen, communityPackTitleField(), "Title", g.packSetupTitle, g.packSetupField == 0)
	drawPublishField(screen, communityPackDescriptionField(), "Description", g.packSetupDescription, g.packSetupField == 1)
	drawButton(screen, communityPackSaveDraftButton(), "Save Draft")
	publishLabel := "Publish"
	if g.packPublishAwaitingID != "" {
		publishLabel = "Publishing..."
	}
	drawButton(screen, communityPackSetupPublishButton(), publishLabel)
}

func drawCommunityBackdrop(screen *ebiten.Image) {
	screen.Fill(colPanel)
	const pixelSize = 6
	elapsed := float64(time.Now().UnixMilli()%240000) / 1000
	for y := 0; y < 186; y += pixelSize {
		for x := 0; x < ScreenWidth; x += pixelSize {
			n := perlin2D(float64(x)/92+elapsed*0.025, float64(y)/72+elapsed*0.009)
			cloud := math.Max(0, math.Abs(n)-0.08) / 0.55
			cloud = math.Min(1, cloud)
			base := color.RGBA{14, 18, 27, 255}
			if n >= 0 {
				base = mixCommunityColor(base, color.RGBA{43, 100, 101, 255}, cloud)
			} else {
				base = mixCommunityColor(base, color.RGBA{116, 62, 80, 255}, cloud)
			}
			vector.DrawFilledRect(screen, float32(x), float32(y), pixelSize, pixelSize, base, false)
			star := communityStarHash(x/pixelSize, y/pixelSize)
			if star%113 == 0 {
				size := float32(2)
				if star%7 == 0 {
					size = 3
				}
				vector.DrawFilledRect(screen, float32(x+2), float32(y+2), size, size, colWhite, false)
			}
		}
	}
	drawCommunityShip(screen, shipPosition(elapsed, 17, 620), 20, false, color.RGBA{239, 235, 220, 255})
	drawCommunityShip(screen, shipPosition(elapsed, 10, 700), 146, true, color.RGBA{235, 107, 86, 255})
	drawCommunityShip(screen, shipPosition(elapsed+31, 7, 760), 118, false, color.RGBA{75, 143, 140, 255})
	vector.DrawFilledRect(screen, 0, 176, ScreenWidth, 12, colGridHeavy, false)
	drawRounded(screen, rect{x: 62, y: 32, w: 416, h: 92}, 8, color.RGBA{45, 45, 43, 255})
	drawRounded(screen, rect{x: 76, y: 46, w: 388, h: 52}, 6, colWhite)
}

func shipPosition(elapsed, speed, distance float64) int {
	return int(math.Mod(elapsed*speed, distance)) - 60
}

func drawCommunityShip(screen *ebiten.Image, x, y int, reverse bool, hull color.RGBA) {
	direction := 1
	if reverse {
		x = ScreenWidth - x
		direction = -1
	}
	px := func(offsetX, offsetY, width, height int, c color.Color) {
		if direction < 0 {
			offsetX = -offsetX - width
		}
		vector.DrawFilledRect(screen, float32(x+offsetX), float32(y+offsetY), float32(width), float32(height), c, false)
	}
	px(-18, -3, 12, 6, color.RGBA{244, 201, 93, 210})
	px(-10, -5, 24, 10, hull)
	px(8, -9, 12, 18, hull)
	px(12, -4, 8, 8, color.RGBA{86, 115, 134, 255})
	px(-2, -10, 8, 5, hull)
	px(-2, 5, 8, 5, hull)
}

func perlin2D(x, y float64) float64 {
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	sx := perlinFade(x - x0)
	sy := perlinFade(y - y0)
	n00 := perlinGradient(int(x0), int(y0), x-x0, y-y0)
	n10 := perlinGradient(int(x0)+1, int(y0), x-x0-1, y-y0)
	n01 := perlinGradient(int(x0), int(y0)+1, x-x0, y-y0-1)
	n11 := perlinGradient(int(x0)+1, int(y0)+1, x-x0-1, y-y0-1)
	return perlinLerp(perlinLerp(n00, n10, sx), perlinLerp(n01, n11, sx), sy)
}

func perlinFade(t float64) float64       { return t * t * t * (t*(t*6-15) + 10) }
func perlinLerp(a, b, t float64) float64 { return a + (b-a)*t }

func perlinGradient(ix, iy int, x, y float64) float64 {
	switch communityStarHash(ix, iy) & 7 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	case 3:
		return -x - y
	case 4:
		return x
	case 5:
		return -x
	case 6:
		return y
	default:
		return -y
	}
}

func communityStarHash(x, y int) uint32 {
	h := uint32(x)*0x8da6b343 ^ uint32(y)*0xd8163841
	h ^= h >> 13
	h *= 0x85ebca6b
	return h ^ (h >> 16)
}

func mixCommunityColor(a, b color.RGBA, amount float64) color.RGBA {
	mix := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*amount) }
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}

func drawLibraryTabs(screen *ebiten.Image, art bool) {
	drawSelectedButton(screen, communityLibraryArtTab(), "Art", art)
	drawSelectedButton(screen, communityLibraryPacksTab(), "Packs", !art)
	drawSelectedButton(screen, communityLibraryPublishedTab(), "Published", false)
}

func (g *Game) drawCommunityPublished(screen *ebiten.Image) {
	drawCenteredText(screen, "MY LIBRARY", rect{x: 100, y: 190, w: 340, h: 30}, colInk)
	drawSelectedButton(screen, communityLibraryArtTab(), "Art", false)
	drawSelectedButton(screen, communityLibraryPacksTab(), "Packs", false)
	drawSelectedButton(screen, communityLibraryPublishedTab(), "Published", true)
	if len(g.communityPublished) == 0 {
		drawCenteredText(screen, "Nothing published yet", rect{x: 70, y: 360, w: 400, h: 32}, colMuted)
		return
	}
	for slot, item := range g.communityPublished {
		if slot >= 4 {
			break
		}
		r := communityPublishedRect(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		if item.Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		} else if len(item.Levels) > 0 && item.Levels[0].Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Levels[0].Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		}
		title := item.Title
		if len(title) > 17 {
			title = title[:17]
		}
		drawText(screen, title, int(r.x+72), int(r.y+24), colInk)
		drawText(screen, item.Kind, int(r.x+72), int(r.y+49), colMuted)
		drawButton(screen, communityPublishedPinButton(slot), "pin")
		drawButton(screen, communityPublishedRemoveButton(slot), "unpublish")
	}
}

func (g *Game) drawCommunityPublishSetup(screen *ebiten.Image) {
	draft, ok := g.communityLibrary.Draft(g.publishDraftID)
	if !ok {
		return
	}
	drawCenteredText(screen, "PUBLISH ART", rect{x: 80, y: 190, w: 380, h: 30}, colInk)
	drawCommunityArtThumbnail(screen, draft.Puzzle.SkeletonRaw, rect{x: 52, y: 238, w: 94, h: 94})
	cover := draft.Puzzle.RevealRaw
	if len(g.publishPreviewRaw) > 0 {
		cover = g.publishPreviewRaw
	}
	drawCommunityArtThumbnail(screen, cover, rect{x: 154, y: 238, w: 94, h: 94})
	drawButton(screen, communityPublishCoverButton(), "Upload cover")
	drawPublishField(screen, communityPublishTitleField(), "Name", g.publishTitle, g.publishField == 0)
	drawPublishField(screen, communityPublishDescriptionField(), "Description", g.publishDescription, g.publishField == 1)
	drawPublishField(screen, communityPublishTagsField(), "Tags", g.publishTags, g.publishField == 2)
	drawSelectedButton(screen, communityPublishOfficialButton(), checkboxLabel(g.publishSubmitOfficial, "Main game review"), g.publishSubmitOfficial)
	if g.publishSubmitOfficial {
		drawCenteredText(screen, "May join the main game after enough", rect{x: 76, y: 470, w: 388, h: 20}, colMuted)
		drawCenteredText(screen, "community upvotes and creator review.", rect{x: 76, y: 488, w: 388, h: 20}, colMuted)
		drawSelectedButton(screen, communityPublishRightsButton(), checkboxLabel(g.publishRightsConfirmed, "I own this art"), g.publishRightsConfirmed)
	}
	publishLabel := "Publish"
	if g.publishAwaitingID != "" {
		publishLabel = "Publishing..."
	}
	drawButton(screen, communityPublishConfirmButton(), publishLabel)
}

func drawPublishField(screen *ebiten.Image, r rect, label, value string, active bool) {
	drawText(screen, label, int(r.x), int(r.y-8), colMuted)
	drawRounded(screen, r, 4, colWhite)
	drawRectOutline(screen, r, 2, colGridHeavy)
	if active {
		drawRounded(screen, inset(r, -3), 7, colAccent)
		drawRounded(screen, r, 4, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
	}
	shown := value
	max := int((r.w - 16) / 8)
	if len(shown) > max {
		shown = shown[len(shown)-max:]
	}
	drawText(screen, shown, int(r.x+8), int(r.y+25), colInk)
}

func checkboxLabel(checked bool, text string) string {
	if checked {
		return "x  " + text
	}
	return "o  " + text
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

func communityBackButton() rect          { return rect{x: 202, y: 674, w: 136, h: 42} }
func communityAccountButton() rect       { return rect{x: 322, y: 208, w: 98, h: 40} }
func communityProfileBadgeButton() rect  { return rect{x: 430, y: 198, w: 58, h: 58} }
func communityLevelsButton() rect        { return rect{x: 108, y: 314, w: 324, h: 50} }
func communityPacksButton() rect         { return rect{x: 128, y: 348, w: 284, h: 46} }
func communityCreateButton() rect        { return rect{x: 128, y: 410, w: 284, h: 46} }
func communityMyArtButton() rect         { return rect{x: 108, y: 388, w: 324, h: 50} }
func communityCreatorsButton() rect      { return rect{x: 108, y: 462, w: 324, h: 50} }
func communityNewButton() rect           { return rect{x: 104, y: 270, w: 332, h: 48} }
func communityImportButton() rect        { return rect{x: 104, y: 338, w: 278, h: 48} }
func communityImportHelpButton() rect    { return rect{x: 392, y: 338, w: 44, h: 48} }
func communityLibraryArtTab() rect       { return rect{x: 54, y: 226, w: 136, h: 36} }
func communityLibraryPacksTab() rect     { return rect{x: 194, y: 226, w: 136, h: 36} }
func communityLibraryPublishedTab() rect { return rect{x: 334, y: 226, w: 152, h: 36} }
func communityPublishedRect(slot int) rect {
	return rect{x: 48, y: 278 + float64(slot)*78, w: 444, h: 70}
}
func communityPublishedPinButton(slot int) rect {
	r := communityPublishedRect(slot)
	return rect{x: r.x + 284, y: r.y + 17, w: 54, h: 36}
}
func communityPublishedRemoveButton(slot int) rect {
	r := communityPublishedRect(slot)
	return rect{x: r.x + 344, y: r.y + 17, w: 88, h: 36}
}
func communityPublishTitleField() rect       { return rect{x: 270, y: 238, w: 220, h: 40} }
func communityPublishDescriptionField() rect { return rect{x: 270, y: 300, w: 220, h: 40} }
func communityPublishTagsField() rect        { return rect{x: 270, y: 362, w: 220, h: 40} }
func communityPublishOfficialButton() rect   { return rect{x: 88, y: 426, w: 364, h: 40} }
func communityPublishRightsButton() rect     { return rect{x: 88, y: 514, w: 364, h: 40} }
func communityPublishConfirmButton() rect    { return rect{x: 170, y: 566, w: 200, h: 44} }
func communityPublishCoverButton() rect      { return rect{x: 52, y: 346, w: 196, h: 38} }
func communityImportPreviewRect(slot int) rect {
	column := slot % 3
	row := slot / 3
	return rect{x: 68 + float64(column)*140, y: 246 + float64(row)*120, w: 104, h: 104}
}
func communityImportConfirmButton() rect { return rect{x: 142, y: 540, w: 256, h: 44} }
func communityArtSearchField() rect      { return rect{x: 48, y: 290, w: 300, h: 32} }
func communityArtCreateButton() rect     { return rect{x: 358, y: 290, w: 134, h: 32} }
func communityNewArtTitleField() rect    { return rect{x: 96, y: 290, w: 348, h: 44} }
func communityNewArtStartButton() rect   { return rect{x: 154, y: 370, w: 232, h: 44} }
func communityPackSetupPreview(art int) rect {
	column := art % 4
	row := art / 4
	return rect{x: 82 + float64(column)*96, y: 238 + float64(row)*64, w: 56, h: 56}
}
func communityPackUploadCoverButton() rect  { return rect{x: 172, y: 342, w: 196, h: 30} }
func communityPackTitleField() rect         { return rect{x: 76, y: 400, w: 388, h: 40} }
func communityPackDescriptionField() rect   { return rect{x: 76, y: 466, w: 388, h: 40} }
func communityPackSaveDraftButton() rect    { return rect{x: 76, y: 536, w: 184, h: 44} }
func communityPackSetupPublishButton() rect { return rect{x: 280, y: 536, w: 184, h: 44} }
func communityDraftRect(slot int) rect {
	return rect{x: 54, y: 234 + float64(slot)*88, w: 432, h: 74}
}
func communityMyArtRect(slot int) rect {
	return rect{x: 48, y: 332 + float64(slot)*92, w: 444, h: 82}
}
func communityMyArtBeforePreviewRect(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 10, y: r.y + 10, w: 62, h: 62}
}
func communityMyArtAfterPreviewRect(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 80, y: r.y + 10, w: 62, h: 62}
}
func communityDraftEditButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 300, y: r.y + 23, w: 40, h: 38}
}
func communityDraftPublishButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 346, y: r.y + 23, w: 46, h: 38}
}
func communityDraftDeleteButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 398, y: r.y + 23, w: 34, h: 38}
}
func communityPrevButton() rect { return rect{x: 62, y: 604, w: 92, h: 38} }
func communityNextButton() rect { return rect{x: 386, y: 604, w: 92, h: 38} }
func communityCatalogPlayButton(slot int) rect {
	r := communityDraftRect(slot)
	return rect{x: r.x + r.w - 90, y: r.y + 17, w: 72, h: 40}
}
func communityGalleryAllButton() rect   { return rect{x: 48, y: 226, w: 62, h: 34} }
func communityGalleryArtButton() rect   { return rect{x: 114, y: 226, w: 62, h: 34} }
func communityGalleryPacksButton() rect { return rect{x: 180, y: 226, w: 78, h: 34} }
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
func communityGoogleButton() rect         { return rect{x: 122, y: 370, w: 296, h: 42} }
func communityEmailInput() rect           { return rect{x: 102, y: 434, w: 336, h: 44} }
func communitySendLinkButton() rect       { return rect{x: 142, y: 500, w: 256, h: 42} }
func communityAccountBioField() rect      { return rect{x: 102, y: 430, w: 336, h: 44} }
func communityAccountBioSaveButton() rect { return rect{x: 102, y: 506, w: 158, h: 42} }
func communitySignOutButton() rect        { return rect{x: 280, y: 506, w: 158, h: 42} }
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

var importExampleBefore = pixelExample([]string{
	"00011000", "00111100", "01111110", "11111111",
	"11111111", "01111110", "00111100", "00011000",
}, []string{"#000000FF"})

var importExampleAfter = pixelExample([]string{
	"00011000", "00122100", "01222210", "12233221",
	"12233221", "01222210", "00122100", "00011000",
}, []string{"#EB6B56FF", "#F4C95DFF", "#4B8F8CFF"})

func pixelExample(rows []string, colors []string) [][]string {
	pixels := make([][]string, len(rows))
	for y, row := range rows {
		pixels[y] = make([]string, len(row))
		for x, value := range row {
			if value == '0' {
				continue
			}
			index := int(value - '1')
			if index >= 0 && index < len(colors) {
				pixels[y][x] = colors[index]
			}
		}
	}
	return pixels
}
