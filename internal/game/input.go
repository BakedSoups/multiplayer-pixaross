package game

import (
	"encoding/json"
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
	if g.pendingPackPublishID != "" && !time.Now().Before(g.pendingPackPublishAt) {
		id := g.pendingPackPublishID
		g.pendingPackPublishID = ""
		g.publishLocalPack(id)
		return
	}
	if g.pendingPublishID != "" && !time.Now().Before(g.pendingPublishAt) {
		id := g.pendingPublishID
		g.pendingPublishID = ""
		g.publishCommunityDraft(id)
		return
	}
	if id := takeCommunityPublishedID(); id != "" {
		g.markCommunityDraftPublished(id)
	}
	if id := takeCommunityPublishedPackID(); id != "" {
		g.markCommunityPackPublished(id)
	}
	if raw := takeCommunityGallery(); raw != "" {
		if err := g.loadCommunityGallery(raw); err != nil {
			g.showCommunityNotice("could not load gallery")
		}
	}
	if raw := takeCommunityPublished(); raw != "" {
		if err := g.loadCommunityPublished(raw); err != nil {
			g.showCommunityNotice("could not load published work")
		}
	}
	if raw := takeCommunityCreators(); raw != "" {
		if err := g.loadCommunityCreators(raw); err != nil {
			g.showCommunityNotice("could not load creators")
		}
	}
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
		g.publishAwaitingID = ""
		g.packPublishAwaitingID = ""
		g.showCommunityNotice(result)
	}
	if raw := takeCommunityImport(); raw != "" {
		if err := g.loadCommunityImportPreview(raw); err != nil {
			g.showCommunityNotice("import failed: " + err.Error())
		}
	}
	if raw := takeCommunityCoverImport(); raw != "" {
		var preview [][]string
		if json.Unmarshal([]byte(raw), &preview) == nil {
			if g.communityView == communityPublishSetup {
				g.publishPreviewRaw = preview
			} else if g.communityView == communityPackSetup {
				g.packSetupPreviewRaw = preview
			}
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
	if g.communityView == communitySignIn && communitySignedIn() && g.profileBioEditing {
		for _, char := range ebiten.AppendInputChars(nil) {
			if char >= 32 && char <= 126 && len(g.profileBioDraft) < 120 {
				g.profileBioDraft += string(char)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.profileBioDraft) > 0 {
			g.profileBioDraft = g.profileBioDraft[:len(g.profileBioDraft)-1]
		}
	}
	if g.communityView == communityPublishSetup {
		for _, char := range ebiten.AppendInputChars(nil) {
			if char < 32 || char > 126 {
				continue
			}
			switch g.publishField {
			case 0:
				if len(g.publishTitle) < 80 {
					g.publishTitle += string(char)
				}
			case 1:
				if len(g.publishDescription) < 160 {
					g.publishDescription += string(char)
				}
			case 2:
				if len(g.publishTags) < 100 {
					g.publishTags += string(char)
				}
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			switch g.publishField {
			case 0:
				if len(g.publishTitle) > 0 {
					g.publishTitle = g.publishTitle[:len(g.publishTitle)-1]
				}
			case 1:
				if len(g.publishDescription) > 0 {
					g.publishDescription = g.publishDescription[:len(g.publishDescription)-1]
				}
			case 2:
				if len(g.publishTags) > 0 {
					g.publishTags = g.publishTags[:len(g.publishTags)-1]
				}
			}
		}
	}
	if g.communityView == communityNewArtSetup {
		for _, char := range ebiten.AppendInputChars(nil) {
			if char >= 32 && char <= 126 && len(g.newArtTitle) < 80 {
				g.newArtTitle += string(char)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.newArtTitle) > 0 {
			g.newArtTitle = g.newArtTitle[:len(g.newArtTitle)-1]
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startNewCommunityArt()
			return
		}
	}
	if g.communityView == communityPackSetup {
		for _, char := range ebiten.AppendInputChars(nil) {
			if char < 32 || char > 126 {
				continue
			}
			if g.packSetupField == 0 && len(g.packSetupTitle) < 80 {
				g.packSetupTitle += string(char)
			}
			if g.packSetupField == 1 && len(g.packSetupDescription) < 200 {
				g.packSetupDescription += string(char)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			if g.packSetupField == 0 && len(g.packSetupTitle) > 0 {
				g.packSetupTitle = g.packSetupTitle[:len(g.packSetupTitle)-1]
			}
			if g.packSetupField == 1 && len(g.packSetupDescription) > 0 {
				g.packSetupDescription = g.packSetupDescription[:len(g.packSetupDescription)-1]
			}
		}
	}
	if g.communityView == communityMyArt && g.artSearchActive {
		changed := false
		for _, char := range ebiten.AppendInputChars(nil) {
			if char >= 32 && char <= 126 && len(g.artSearch) < 60 {
				g.artSearch += string(char)
				changed = true
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.artSearch) > 0 {
			g.artSearch = g.artSearch[:len(g.artSearch)-1]
			changed = true
		}
		if changed {
			g.communityPage = 0
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
		g.profileBioDraft = g.profileBio
		g.communityView = communitySignIn
		return
	}
	switch g.communityView {
	case communityHome:
		switch {
		case communityLevelsButton().Contains(x, y):
			g.communityView = communityBrowse
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
		case communityMyArtButton().Contains(x, y):
			g.communityView = communityMyArt
			g.communityPage = 0
			g.syncLocalDrafts()
			requestCommunityCloudDrafts()
		case communityCreatorsButton().Contains(x, y):
			g.communityView = communityCreators
			g.selectedCreator = -1
			g.communityPage = 0
			g.syncCommunityProfileArt()
			requestCommunityCreators()
		}
	case communityCreate:
		switch {
		case communityNewButton().Contains(x, y):
			g.openNewArtSetup()
		case communityImportButton().Contains(x, y):
			if !requestCommunityImport() {
				g.showCommunityNotice("import is available in the web build")
			}
		case communityImportHelpButton().Contains(x, y):
			g.communityView = communityImportHelp
		}
	case communityMyArt:
		if communityArtSearchField().Contains(x, y) {
			g.artSearchActive = true
			return
		}
		g.artSearchActive = false
		if communityLibraryPublishedTab().Contains(x, y) {
			g.communityView = communityPublished
			g.communityPage = 0
			requestCommunityPublished()
			return
		}
		if communityLibraryPacksTab().Contains(x, y) {
			g.communityView = communityPacks
			g.communityPage = 0
			return
		}
		if communityArtCreateButton().Contains(x, y) {
			g.communityView = communityCreate
			return
		}
		indexes := g.filteredCommunityDraftIndexes()
		start := g.communityPage * communityDraftsPerPage
		for slot := 0; slot < communityDraftsPerPage; slot++ {
			position := start + slot
			if position >= len(indexes) {
				break
			}
			index := indexes[position]
			if communityDraftEditButton(slot).Contains(x, y) {
				g.editCommunityDraft(index)
				return
			}
			if communityDraftPublishButton(slot).Contains(x, y) {
				g.queueCommunityDraftPublish(index)
				return
			}
			if communityDraftDeleteButton(slot).Contains(x, y) {
				g.deleteCommunityDraft(index)
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityDraftsPerPage < len(indexes) {
			g.communityPage++
		}
	case communityBrowse:
		switch {
		case communityGalleryAllButton().Contains(x, y):
			g.galleryKind = "all"
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGalleryArtButton().Contains(x, y):
			g.galleryKind = "art"
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGalleryPacksButton().Contains(x, y):
			g.galleryKind = "pack"
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGalleryNewButton().Contains(x, y):
			g.gallerySort = "new"
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGalleryTopButton().Contains(x, y):
			g.gallerySort = "top"
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		}
		start := g.communityPage * communityCatalogPerPage
		for slot := 0; slot < communityCatalogPerPage; slot++ {
			index := start + slot
			if index >= len(g.communityGallery) {
				continue
			}
			item := g.communityGallery[index]
			if communityGalleryOpenButton(slot).Contains(x, y) {
				if item.Kind == "pack" {
					g.selectedGallery = index
					g.communityView = communityGalleryPack
				} else {
					g.playGalleryLevel(index)
				}
				return
			}
			if communityGalleryLikeButton(slot).Contains(x, y) {
				if !communitySignedIn() {
					g.communityView = communitySignIn
				} else {
					toggleCommunityLike(item.Kind, item.ID)
				}
				return
			}
			if item.Owned && communityGalleryPromoteButton(slot).Contains(x, y) {
				promoteCommunityItem(item.Kind, item.ID)
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityCatalogPerPage < len(g.communityGallery) {
			g.communityPage++
		}
	case communityGalleryPack:
		if g.selectedGallery >= 0 && g.selectedGallery < len(g.communityGallery) {
			for slot := 0; slot < 6 && slot < len(g.communityGallery[g.selectedGallery].Levels); slot++ {
				if communityGalleryPackLevelButton(slot).Contains(x, y) {
					g.playGalleryPackLevel(slot)
					return
				}
			}
		}
	case communityPacks:
		if communityLibraryPublishedTab().Contains(x, y) {
			g.communityView = communityPublished
			g.communityPage = 0
			requestCommunityPublished()
			return
		}
		if communityLibraryArtTab().Contains(x, y) {
			g.communityView = communityMyArt
			g.communityPage = 0
			return
		}
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
				g.queueLocalPackPublish(slot)
				return
			}
			if communityPackDeleteButton(slot).Contains(x, y) {
				g.deleteCommunityPack(slot)
				return
			}
		}
	case communityPublished:
		if communityLibraryArtTab().Contains(x, y) {
			g.communityView = communityMyArt
			g.communityPage = 0
			return
		}
		if communityLibraryPacksTab().Contains(x, y) {
			g.communityView = communityPacks
			g.communityPage = 0
			return
		}
		for slot, item := range g.communityPublished {
			if slot >= 4 {
				break
			}
			if communityPublishedPinButton(slot).Contains(x, y) {
				promoteCommunityItem(item.Kind, item.ID)
				return
			}
			if communityPublishedRemoveButton(slot).Contains(x, y) {
				unpublishCommunityItem(item.Kind, item.ID)
				return
			}
		}
	case communityPublishSetup:
		switch {
		case communityPublishCoverButton().Contains(x, y):
			requestCommunityCoverImport(10)
		case communityPublishTitleField().Contains(x, y):
			g.publishField = 0
		case communityPublishDescriptionField().Contains(x, y):
			g.publishField = 1
		case communityPublishTagsField().Contains(x, y):
			g.publishField = 2
		case communityPublishOfficialButton().Contains(x, y):
			g.publishSubmitOfficial = !g.publishSubmitOfficial
			if !g.publishSubmitOfficial {
				g.publishRightsConfirmed = false
			}
		case communityPublishRightsButton().Contains(x, y) && g.publishSubmitOfficial:
			g.publishRightsConfirmed = !g.publishRightsConfirmed
		case communityPublishConfirmButton().Contains(x, y):
			if g.publishAwaitingID == "" {
				g.submitCommunityDraftPublish()
			}
		}
	case communityImportPreview:
		if communityImportConfirmButton().Contains(x, y) {
			if err := g.importCommunityPack(g.communityImportRaw); err != nil {
				g.showCommunityNotice("import failed: " + err.Error())
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
			g.openNewPackSetup()
			return
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
			g.communityPage++
		}
	case communityNewArtSetup:
		if communityNewArtTitleField().Contains(x, y) {
			return
		}
		if communityNewArtStartButton().Contains(x, y) {
			g.startNewCommunityArt()
		}
	case communityPackSetup:
		if communityPackUploadCoverButton().Contains(x, y) {
			requestCommunityCoverImport(10)
			return
		}
		for art := 0; art < len(g.packSetupItems) && art < 8; art++ {
			if communityPackSetupPreview(art).Contains(x, y) {
				g.packSetupPreview = art
				return
			}
		}
		switch {
		case communityPackTitleField().Contains(x, y):
			g.packSetupField = 0
		case communityPackDescriptionField().Contains(x, y):
			g.packSetupField = 1
		case communityPackSaveDraftButton().Contains(x, y):
			g.savePackSetup(false)
		case communityPackSetupPublishButton().Contains(x, y):
			if g.packPublishAwaitingID == "" {
				g.savePackSetup(true)
			}
		}
	case communitySignIn:
		if communitySignedIn() {
			switch {
			case communityAccountBioField().Contains(x, y):
				g.profileBioEditing = true
			case communityAccountBioSaveButton().Contains(x, y):
				g.profileBio = strings.TrimSpace(g.profileBioDraft)
				saveCommunityBio(g.profileBio)
				if raw, err := json.Marshal(g.profileArt.puzzle()); err == nil {
					syncCommunityProfile(string(raw), g.profileBio)
				}
				g.profileBioEditing = false
			case communitySignOutButton().Contains(x, y):
				requestCommunitySignOut()
				g.communityView = communityHome
			}
		} else {
			switch {
			case communityGoogleButton().Contains(x, y):
				if !requestCommunityGoogleSignIn() {
					g.showCommunityNotice("Google sign in is unavailable")
				}
			case communitySendLinkButton().Contains(x, y):
				g.submitCommunitySignIn()
			}
		}
	case communityCreators:
		start := g.communityPage * 4
		for slot := 0; slot < 4 && start+slot < len(g.communityCreators); slot++ {
			if communityCreatorButton(slot).Contains(x, y) {
				g.selectedCreator = start + slot
				g.communityPage = 0
				g.communityView = communityCreatorProfile
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*4 < len(g.communityCreators) {
			g.communityPage++
		}
	case communityCreatorProfile:
		if g.selectedCreator >= 0 && g.selectedCreator < len(g.communityCreators) {
			levels := g.communityCreators[g.selectedCreator].Levels
			contentY := float64(310)
			if len(g.communityCreators[g.selectedCreator].Featured) > 0 {
				contentY = 410
			}
			start := g.communityPage * 4
			for slot := 0; slot < 4 && start+slot < len(levels); slot++ {
				if communityCreatorLevelButtonAt(slot, contentY).Contains(x, y) {
					g.playCreatorLevel(start + slot)
					return
				}
			}
			if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
				g.communityPage--
			}
			if communityNextButton().Contains(x, y) && (g.communityPage+1)*4 < len(levels) {
				g.communityPage++
			}
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
	if g.communityView == communityNewArtSetup {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityPackSetup {
		if g.packSetupID == "" {
			g.communityView = communityPackBuild
		} else {
			g.communityView = communityPacks
		}
		return
	}
	if g.communityView == communityPublishSetup {
		g.communityView = communityMyArt
		return
	}
	if g.communityView == communityImportPreview {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityImportHelp {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityCreate {
		g.communityView = communityMyArt
		return
	}
	if g.communityView == communityGalleryPack {
		g.communityView = communityBrowse
		return
	}
	if g.communityView == communityCreatorProfile {
		g.communityView = communityCreators
		return
	}
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
		g.communityView = communityMyArt
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
			return true
		case profileSaveButton().Contains(x, y):
			g.closeProfileEditor(true)
			return true
		}
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
