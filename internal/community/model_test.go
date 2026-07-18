package community

import (
	"testing"

	"github.com/alex/nongrampictures/internal/nonogram"
)

func TestDraftPublishValidation(t *testing.T) {
	puzzle := &nonogram.Puzzle{
		ID:          "draft",
		Title:       "Flower",
		Width:       8,
		Height:      8,
		SolutionRaw: rows(8, "10000000"),
		SkeletonRaw: pixels(8, "#000000FF"),
		RevealRaw:   pixels(8, "#975347FF"),
	}
	draft := NewDraft("level-1", puzzle)
	if err := draft.ValidateForPublish(); err != nil {
		t.Fatal(err)
	}
}

func TestBeforeLayerMustBeBlack(t *testing.T) {
	puzzle := &nonogram.Puzzle{
		ID:          "draft",
		Title:       "Flower",
		Width:       8,
		Height:      8,
		SolutionRaw: rows(8, "10000000"),
		SkeletonRaw: pixels(8, "#FF0000FF"),
		RevealRaw:   pixels(8, "#975347FF"),
	}
	draft := NewDraft("level-1", puzzle)
	if err := draft.ValidateForSave(); err == nil {
		t.Fatal("colored before layer passed validation")
	}
}

func rows(count int, first string) []string {
	result := make([]string, count)
	result[0] = first
	for i := 1; i < count; i++ {
		result[i] = "00000000"
	}
	return result
}

func pixels(size int, first string) [][]string {
	result := make([][]string, size)
	for y := range result {
		result[y] = make([]string, size)
	}
	result[0][0] = first
	return result
}
