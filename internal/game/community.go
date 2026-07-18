package game

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alex/nongrampictures/internal/assets"
	"github.com/alex/nongrampictures/internal/community"
	"github.com/alex/nongrampictures/internal/nonogram"
)

type communityView uint8

const (
	communityHome communityView = iota
	communityBrowse
	communityPacks
	communityCreate
	communityMyArt
	communitySignIn
	communityPackBuild
	communityCreators
	communityCreatorProfile
)

func loadCommunityLibrary() community.Library {
	var library community.Library
	raw := loadCommunityData()
	if raw == "" || json.Unmarshal([]byte(raw), &library) != nil {
		return community.Library{}
	}
	return library
}

func (g *Game) saveCommunityLibrary() bool {
	raw, err := json.Marshal(g.communityLibrary)
	return err == nil && saveCommunityData(string(raw))
}

func (g *Game) newCommunityDraft(size int) {
	g.editor = newEditorState(size)
	g.editor.Title = "Untitled"
	g.currentDraftID = newLocalID("level")
	g.editorUndo = nil
	g.editorSizeOpen = false
	g.editorOnionSkin = false
	g.mode = screenEditor
	_ = g.saveCurrentDraft(false)
}

func (g *Game) openProfileEditor() {
	g.profileReturn = g.editor.clone()
	g.profileDraftID = g.currentDraftID
	g.editor = g.profileArt.clone()
	g.editor.Title = "Profile"
	g.editor.selectLayer(editorLayerAfter)
	g.editorUndo = nil
	g.editorSizeOpen = false
	g.editorOnionSkin = false
	g.editingProfile = true
	g.mode = screenEditor
}

func (g *Game) closeProfileEditor(save bool) {
	if save {
		g.editor.Title = "Profile"
		g.editor.selectLayer(editorLayerAfter)
		g.profileArt = g.editor.clone()
		saveCommunityProfile(g.profileArt.packJSON())
		if raw, err := json.Marshal(g.profileArt.puzzle()); err == nil {
			syncCommunityProfile(string(raw))
		}
	}
	g.editor = g.profileReturn
	g.currentDraftID = g.profileDraftID
	g.editingProfile = false
	g.editorUndo = nil
	g.communityView = communityHome
	g.mode = screenCommunity
}

func (g *Game) editCommunityDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	draft := g.communityLibrary.Drafts[index]
	raw, err := json.Marshal(editorPack{Levels: []*nonogram.Puzzle{draft.Puzzle}})
	if err != nil {
		return
	}
	editor, err := editorFromPackJSON(string(raw))
	if err != nil {
		g.showCommunityNotice("draft could not be opened")
		return
	}
	g.editor = editor
	g.editor.Title = draft.Title
	g.currentDraftID = draft.ID
	g.editorUndo = nil
	g.mode = screenEditor
}

func (g *Game) saveCurrentDraft(playtested bool) error {
	puzzle := g.editor.puzzle()
	id := g.currentDraftID
	if id == "" {
		id = newLocalID("level")
		g.currentDraftID = id
	}
	puzzle.ID = id
	title := strings.TrimSpace(g.editor.Title)
	if title == "" {
		title = "Untitled"
	}
	puzzle.Title = title
	draft := community.NewDraft(id, puzzle)
	if existing, ok := g.communityLibrary.Draft(id); ok {
		draft.OwnerID = existing.OwnerID
		draft.Description = existing.Description
		draft.Tags = append([]string(nil), existing.Tags...)
		draft.Visibility = existing.Visibility
		draft.Status = existing.Status
		draft.Version = existing.Version
		draft.Playtested = existing.Playtested || playtested
		draft.Stats = existing.Stats
	}
	if err := draft.ValidateForSave(); err != nil {
		return err
	}
	g.communityLibrary.UpsertDraft(draft)
	if !g.saveCommunityLibrary() {
		return fmt.Errorf("local storage is unavailable")
	}
	if raw, err := json.Marshal(draft); err == nil {
		syncCommunityDraft(string(raw))
	}
	_ = saveEditorPack(g.editor.packJSON())
	return nil
}

func (g *Game) mergeCloudDrafts(raw string) error {
	var drafts []community.LevelDraft
	if err := json.Unmarshal([]byte(raw), &drafts); err != nil {
		return err
	}
	for _, draft := range drafts {
		if draft.Puzzle == nil || draft.ValidateForSave() != nil {
			continue
		}
		existing, ok := g.communityLibrary.Draft(draft.ID)
		if !ok || existing.UpdatedAt < draft.UpdatedAt {
			g.communityLibrary.UpsertDraft(draft)
		}
	}
	g.saveCommunityLibrary()
	return nil
}

func (g *Game) syncLocalDrafts() {
	for _, draft := range g.communityLibrary.Drafts {
		if raw, err := json.Marshal(draft); err == nil {
			syncCommunityDraft(string(raw))
		}
	}
}

func (g *Game) importCommunityPack(raw string) error {
	var pack editorPack
	if err := json.Unmarshal([]byte(raw), &pack); err != nil {
		return err
	}
	if len(pack.Levels) == 0 {
		return fmt.Errorf("no paired levels were found")
	}
	imported := 0
	for _, puzzle := range pack.Levels {
		if puzzle == nil {
			continue
		}
		if err := puzzle.ParseSolution(); err != nil {
			continue
		}
		id := newLocalID(fmt.Sprintf("import-%d", imported+1))
		puzzle.ID = id
		draft := community.NewDraft(id, puzzle)
		if err := draft.ValidateForSave(); err != nil {
			continue
		}
		g.communityLibrary.UpsertDraft(draft)
		imported++
	}
	if imported == 0 {
		return fmt.Errorf("no valid levels were found")
	}
	if !g.saveCommunityLibrary() {
		return fmt.Errorf("local storage is unavailable")
	}
	g.communityView = communityMyArt
	g.communityPage = 0
	g.showCommunityNotice(fmt.Sprintf("imported %d level(s)", imported))
	return nil
}

func (g *Game) publishCommunityDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	draft := g.communityLibrary.Drafts[index]
	if err := draft.ValidateForPublish(); err != nil {
		g.showCommunityNotice(err.Error())
		return
	}
	if !communitySignedIn() {
		g.communityView = communitySignIn
		g.showCommunityNotice("sign in to publish")
		return
	}
	raw, err := json.Marshal(draft)
	if err != nil || !requestCommunityPublish(string(raw)) {
		g.showCommunityNotice("publishing is available in the web build")
	}
}

func (g *Game) loadCommunityCatalog(raw string) error {
	var versions []community.LevelVersion
	if err := json.Unmarshal([]byte(raw), &versions); err != nil {
		return err
	}
	for i := range versions {
		if versions[i].Puzzle != nil {
			if err := versions[i].Puzzle.ParseSolution(); err != nil {
				return err
			}
		}
	}
	g.communityCatalog = versions
	return nil
}

func (g *Game) loadCommunityCreators(raw string) error {
	var creators []community.CreatorProfile
	if err := json.Unmarshal([]byte(raw), &creators); err != nil {
		return err
	}
	for i := range creators {
		if creators[i].AvatarPuzzle != nil {
			_ = creators[i].AvatarPuzzle.ParseSolution()
		}
		for j := range creators[i].Levels {
			if creators[i].Levels[j].Puzzle != nil {
				if err := creators[i].Levels[j].Puzzle.ParseSolution(); err != nil {
					return err
				}
			}
		}
	}
	g.communityCreators = creators
	return nil
}

func (g *Game) playCreatorLevel(index int) {
	if g.selectedCreator < 0 || g.selectedCreator >= len(g.communityCreators) {
		return
	}
	levels := g.communityCreators[g.selectedCreator].Levels
	if index < 0 || index >= len(levels) || levels[index].Puzzle == nil {
		return
	}
	g.activeCommunityPack = ""
	g.communityPlayReturn = communityCreatorProfile
	g.loadCommunityPuzzle(levels[index].Puzzle)
}

func (g *Game) playCommunityVersion(index int) {
	if index < 0 || index >= len(g.communityCatalog) || g.communityCatalog[index].Puzzle == nil {
		return
	}
	p := g.communityCatalog[index].Puzzle
	g.activeCommunityPack = ""
	g.communityPlayReturn = communityBrowse
	g.loadCommunityPuzzle(p)
}

func (g *Game) loadCommunityPuzzle(p *nonogram.Puzzle) {
	g.puzzle = p
	g.board = nonogram.NewBoard(p.Width, p.Height)
	g.rowClues = nonogram.RowClues(p.Solution)
	g.colClues = nonogram.ColumnClues(p.Solution)
	g.skeleton = nil
	g.reveal = nil
	g.skeletonPixels = pixelsFromRaw(p.SkeletonRaw)
	g.revealPixels = pixelsFromRaw(p.RevealRaw)
	g.undoStack = nil
	g.startTime = time.Now()
	g.editorPreview = false
	g.communityPreview = true
	g.mode = screenPuzzle
}

func pixelsFromRaw(raw [][]string) [][]assets.PixelCell {
	result := make([][]assets.PixelCell, len(raw))
	for y, row := range raw {
		result[y] = make([]assets.PixelCell, len(row))
		for x, value := range row {
			if c, ok := parseEditorHexColor(value); ok && c.A > 0 {
				result[y][x] = assets.PixelCell{Color: c, Visible: true}
			}
		}
	}
	return result
}

func (g *Game) createLocalPack() {
	items := make([]community.PackItem, 0, len(g.packSelection))
	for _, draft := range g.communityLibrary.Drafts {
		if !g.packSelection[draft.ID] {
			continue
		}
		items = append(items, community.PackItem{LevelID: draft.ID, Position: len(items)})
		if len(items) == 20 {
			break
		}
	}
	pack := community.Pack{
		ID:         newLocalID("pack"),
		Title:      fmt.Sprintf("My Pack %d", len(g.communityLibrary.Packs)+1),
		Visibility: community.VisibilityDraft,
		Status:     community.LevelDraftStatus,
		Version:    1,
		Items:      items,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	if err := pack.Validate(); err != nil {
		g.showCommunityNotice("create at least one level first")
		return
	}
	g.communityLibrary.Packs = append([]community.Pack{pack}, g.communityLibrary.Packs...)
	g.saveCommunityLibrary()
	g.packSelection = nil
	g.communityView = communityPacks
	g.showCommunityNotice("pack created")
}

func (g *Game) openPackBuilder() {
	g.packSelection = make(map[string]bool)
	g.communityPage = 0
	g.communityView = communityPackBuild
}

func (g *Game) togglePackDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	id := g.communityLibrary.Drafts[index].ID
	if !g.packSelection[id] && len(g.packSelection) >= 20 {
		g.showCommunityNotice("packs can contain up to 20 levels")
		return
	}
	if g.packSelection[id] {
		delete(g.packSelection, id)
	} else {
		g.packSelection[id] = true
	}
}

func (g *Game) publishLocalPack(index int) {
	if index < 0 || index >= len(g.communityLibrary.Packs) {
		return
	}
	pack := g.communityLibrary.Packs[index]
	if err := pack.Validate(); err != nil {
		g.showCommunityNotice(err.Error())
		return
	}
	if !communitySignedIn() {
		g.communityView = communitySignIn
		g.showCommunityNotice("sign in to publish")
		return
	}
	raw, err := json.Marshal(pack)
	if err != nil || !requestCommunityPackPublish(string(raw)) {
		g.showCommunityNotice("pack publishing is available in the web build")
	}
}

func (g *Game) playLocalPack(index int) {
	if index < 0 || index >= len(g.communityLibrary.Packs) {
		return
	}
	pack := &g.communityLibrary.Packs[index]
	completed := make(map[string]bool, len(pack.Progress.CompletedLevelIDs))
	for _, id := range pack.Progress.CompletedLevelIDs {
		completed[id] = true
	}
	for _, item := range pack.Items {
		if completed[item.LevelID] {
			continue
		}
		if draft, ok := g.communityLibrary.Draft(item.LevelID); ok && draft.Puzzle != nil {
			g.activeCommunityPack = pack.ID
			g.loadCommunityPuzzle(draft.Puzzle)
			return
		}
	}
	g.showCommunityNotice("pack complete")
}

func (g *Game) completeCommunityPackLevel(levelID string) {
	for i := range g.communityLibrary.Packs {
		pack := &g.communityLibrary.Packs[i]
		if pack.ID != g.activeCommunityPack {
			continue
		}
		for _, id := range pack.Progress.CompletedLevelIDs {
			if id == levelID {
				return
			}
		}
		pack.Progress.CompletedLevelIDs = append(pack.Progress.CompletedLevelIDs, levelID)
		g.saveCommunityLibrary()
		return
	}
}

func newLocalID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (g *Game) showCommunityNotice(text string) {
	g.communityNotice = text
	g.communityNoticeUntil = time.Now().Add(2200 * time.Millisecond)
}
