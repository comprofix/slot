package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"slot/machine"
)

// The game is saved after every change, so it can be closed at any moment
// and picked up where it was left: credits, bet, jackpot pots and all.

// savePath is where the game is kept: a file in the user's own settings
// folder, so it survives the program being moved or rebuilt.
func savePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PokerMachine", "save.json"), nil
}

// loadGame reads the saved game. It reports false if there is none, or if
// it cannot be read: the game then starts afresh.
func loadGame(path string) (machine.Saved, bool) {
	var s machine.Saved
	data, err := os.ReadFile(path)
	if err != nil {
		return s, false
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, false
	}
	return s, true
}

// saveGame writes the saved game. It writes to a second file and then
// renames it, so a crash part-way through cannot spoil the save there was.
func saveGame(path string, s machine.Saved) error {
	data, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// save keeps the game as it stands. A failure is reported on screen once,
// not on every spin.
func (g *Game) save() {
	if g.savePath == "" {
		return
	}
	if err := saveGame(g.savePath, g.m.Save()); err != nil && !g.saveFailed {
		g.saveFailed = true
		g.message = "COULD NOT SAVE THE GAME"
	}
}
