package mysql

import (
	"testing"

	"github.com/a-h/callme/mysql/migrations"
)

func TestThatBinDataIsPresent(t *testing.T) {
	names := migrations.AssetNames()
	if len(names) == 0 {
		t.Errorf("expected more than 0 names, but got %v", len(names))
	}
}

func TestThatBinDataCanBeAccessed(t *testing.T) {
	name := migrations.AssetNames()[0]
	_, err := migrations.Asset(name)
	if err != nil {
		t.Errorf("couldn't load asset '%v' with error '%v'", name, err)
	}
}

func TestMigrationsCanBeLoaded(t *testing.T) {
	driver, err := LoadMigrations()
	if err != nil {
		t.Errorf("failed to load migrations: %v", err)
	}
	defer driver.Close()

	v, err := driver.First()
	if err != nil {
		t.Errorf("failed to find any migrations: %v", err)
	}
	if v == 0 {
		t.Error("got version 0, should have been different")
	}
}
