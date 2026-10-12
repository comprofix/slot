package main

import (
	"os"
	"path/filepath"
	"testing"

	"slot/machine"
)

func TestSaveAndLoadGame(t *testing.T) {
	// The folder does not exist yet: saving has to make it.
	path := filepath.Join(t.TempDir(), "PokerMachine", "save.json")

	if _, ok := loadGame(path); ok {
		t.Error("loaded a game before one was saved")
	}

	m := machine.New(1, 2)
	m.Lines, m.BetPerLine = 25, 2
	m.Spin()
	for m.Link != nil {
		m.Respin()
	}
	want := m.Save()
	if err := saveGame(path, want); err != nil {
		t.Fatal(err)
	}
	got, ok := loadGame(path)
	if !ok {
		t.Fatal("could not load the game just saved")
	}
	if got != want {
		t.Errorf("loaded %+v, want %+v", got, want)
	}

	// Saving again replaces the file and leaves nothing else behind.
	want.Credits += 500
	if err := saveGame(path, want); err != nil {
		t.Fatal(err)
	}
	if got, _ := loadGame(path); got.Credits != want.Credits {
		t.Errorf("loaded %d credits after a second save, want %d", got.Credits, want.Credits)
	}
	files, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("%d files in the save folder, want 1", len(files))
	}

	// A damaged file is ignored rather than stopping the game.
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := loadGame(path); ok {
		t.Error("loaded a damaged save")
	}
}
