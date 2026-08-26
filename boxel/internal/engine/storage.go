package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	p := filepath.Join(dir, "boxel")
	_ = os.MkdirAll(p, 0o755)
	return p
}

func SaveGameLocal(game SavedGame) error {
	b, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir(), "save.json"), b, 0o644)
}

func LoadGameLocal() (SavedGame, bool) {
	b, err := os.ReadFile(filepath.Join(configDir(), "save.json"))
	if err != nil {
		return SavedGame{}, false
	}
	var g SavedGame
	if json.Unmarshal(b, &g) != nil || g.Version != 1 || g.Blocks == nil {
		return SavedGame{}, false
	}
	return g, true
}

func ExportGame(game SavedGame, path string) error {
	if path == "" {
		name := strings.ToLower(strings.ReplaceAll(game.Name, " ", "-"))
		path = filepath.Join(configDir(), name+".boxel.json")
	}
	b, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func ReadGameFile(path string) (SavedGame, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return SavedGame{}, err
	}
	var g SavedGame
	if err := json.Unmarshal(b, &g); err != nil {
		return SavedGame{}, err
	}
	if g.Version != 1 || g.Blocks == nil {
		return SavedGame{}, os.ErrInvalid
	}
	return g, nil
}

type Profile struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	FriendCode string `json:"friendCode"`
}

func SaveProfile(p Profile) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir(), "profile.json"), b, 0o644)
}

func LoadProfileFile() (Profile, bool) {
	b, err := os.ReadFile(filepath.Join(configDir(), "profile.json"))
	if err != nil {
		return Profile{}, false
	}
	var p Profile
	if json.Unmarshal(b, &p) != nil || p.ID == "" {
		return Profile{}, false
	}
	return p, true
}

func LastExportPath(game SavedGame) string {
	name := strings.ToLower(strings.ReplaceAll(game.Name, " ", "-"))
	return filepath.Join(configDir(), name+".boxel.json")
}
