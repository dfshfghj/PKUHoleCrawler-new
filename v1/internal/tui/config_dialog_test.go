package tui

import (
	"reflect"
	"testing"

	"treehole/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigDialogUpdateSwitchesSections(t *testing.T) {
	dialog := NewConfigDialog(nil)
	if dialog.ActiveSection() != ConfigSectionAuth {
		t.Fatalf("initial section = %v, want auth", dialog.ActiveSection())
	}

	dialog.Update(tea.KeyMsg{Type: tea.KeyRight})
	if dialog.ActiveSection() != ConfigSectionDatabase {
		t.Fatalf("section after right = %v, want database", dialog.ActiveSection())
	}
	if dialog.FocusIndex() != 0 {
		t.Fatalf("focus after section switch = %d, want 0", dialog.FocusIndex())
	}

	dialog.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if dialog.ActiveSection() != ConfigSectionAuth {
		t.Fatalf("section after left = %v, want auth", dialog.ActiveSection())
	}
}

func TestConfigDialogToConfigPreservesExistingAndAppliesDatabaseEdits(t *testing.T) {
	existing := &config.Config{
		Username:   "old-user",
		Password:   "old-pass",
		SecretKey:  "old-secret",
		DeviceUUID: "old-device",
		Database: config.DatabaseConfig{
			Type:    "sqlite3",
			DBFile:  "./treehole.db",
			SSLMode: "disable",
		},
		Cors: config.CorsConfig{
			AllowOrigins: []string{"http://localhost:3000"},
			AllowMethods: []string{"GET", "POST"},
			AllowHeaders: []string{"Authorization"},
		},
	}

	dialog := NewConfigDialog(existing)
	dialog.authInputs[0].SetValue("new-user")
	dialog.authInputs[3].SetValue("device-2")
	dialog.databaseInputs[0].SetValue("postgres")
	dialog.databaseInputs[1].SetValue("db.internal")
	dialog.databaseInputs[2].SetValue("15432")
	dialog.databaseInputs[3].SetValue("dbuser")
	dialog.databaseInputs[4].SetValue("dbpass")
	dialog.databaseInputs[5].SetValue("treehole")
	dialog.databaseInputs[6].SetValue("./ignored.db")
	dialog.databaseInputs[7].SetValue("require")
	dialog.databaseInputs[8].SetValue("postgres://example")

	got := dialog.ToConfig(existing)

	if got.Username != "new-user" {
		t.Fatalf("username = %q, want %q", got.Username, "new-user")
	}
	if got.DeviceUUID != "device-2" {
		t.Fatalf("device uuid = %q, want %q", got.DeviceUUID, "device-2")
	}
	if got.Database.Type != "postgres" || got.Database.Host != "db.internal" || got.Database.Port != 15432 {
		t.Fatalf("database not updated correctly: %+v", got.Database)
	}
	if !reflect.DeepEqual(got.Cors, existing.Cors) {
		t.Fatalf("cors changed unexpectedly: got %+v want %+v", got.Cors, existing.Cors)
	}
}
