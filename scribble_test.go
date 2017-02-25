package scribble_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/golang-scribble"
)

//
type Fish struct {
	Type string `json:"type"`
}

//
var (
	db         *scribble.Driver
	database   = "/tmp/deep/school"
	collection = "fish"
	redfish    = Fish{Type: "red"}
	bluefish   = Fish{Type: "blue"}
)

//
func TestMain(m *testing.M) {

	// remove any thing for a potentially failed previous test
	os.RemoveAll("/tmp/deep")

	// run
	code := m.Run()

	// cleanup
	os.RemoveAll("/tmp/deep")

	// exit
	os.Exit(code)
}

// Tests creating a new database, and using an existing database
func TestNew(t *testing.T) {

	// database should not exist
	if _, err := os.Stat(database); err == nil {
		t.Error("Expected nothing, got database")
	}
	logger := lumber.NewConsoleLogger(lumber.WARN)
	// test options
	if _, err := scribble.New(database, &scribble.Options{logger}); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test safe db location
	if err := createDB("/home/fakeuser"); err == nil {
		t.Error("Expected error, got nothing")
		t.FailNow()
	}

	// test uncreatable
	if err := createDB("/root/noway/jose"); err == nil {
		t.Error("Expected error, got nothing")
		t.FailNow()
	}

	// create a new database
	if err := createDB(database); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// database should exist
	if _, err := os.Stat(database); err != nil {
		t.Error("Expected database, got nothing")
		t.FailNow()
	}

	// should use existing database
	createDB(database)

	// database should exist
	if _, err := os.Stat(database); err != nil {
		t.Error("Expected database, got nothing")
		t.FailNow()
	}
}

//
func TestWriteAndRead(t *testing.T) {

	createDB(database)
	defer destroySchool()

	// add fish to database
	if err := db.Write(collection, "redfish", redfish); err != nil {
		t.Error("Create fish failed: ", err.Error())
		t.FailNow()
	}

	// read fish from database
	var onefish Fish
	if err := db.Read(collection, "redfish", &onefish); err != nil {
		t.Error("Failed to read: ", err.Error())
		t.FailNow()
	}

	// ensure unmarshalling went well
	if onefish.Type != redfish.Type {
		t.Error("Expected red fish, got: ", onefish.Type)
		t.FailNow()
	}
}

//
func TestReadall(t *testing.T) {
	createDB(database)
	createSchool()
	defer destroySchool()

	// read all into []string
	records, err := db.ReadAll(collection)
	if err != nil {
		t.Error("Failed to read: ", err.Error())
		t.FailNow()
	}

	if len(records) != 2 {
		t.Errorf("Expected two fishies, have %d", len(records))
		t.FailNow()
	}

	fishies := []Fish{}
	for i := range records {
		fish := Fish{}
		if err = json.Unmarshal([]byte(records[i]), &fish); err != nil {
			t.Errorf("Failed to unmarshal fish - %s", err.Error())
			t.FailNow()
		}
		fishies = append(fishies, fish)
	}

	if fishies[0].Type != redfish.Type && fishies[1].Type != bluefish.Type {
		t.Errorf("Read failed, possibly bad order, got %s - %s", fishies, err.Error())
		t.FailNow()
	}
}

func TestReadallMap(t *testing.T) {
	createDB(database)
	createSchool()
	defer destroySchool()

	// read all into []string
	records, err := db.ReadAllMap(collection)
	if err != nil {
		t.Error("Failed to read: ", err.Error())
		t.FailNow()
	}

	if len(records) != 2 {
		t.Errorf("Expected two fishies, have %d", len(records))
		t.FailNow()
	}

	fishies := map[string]Fish{}
	for i := range records {
		fish := Fish{}
		if err = json.Unmarshal([]byte(records[i]), &fish); err != nil {
			t.Errorf("Failed to unmarshal fish - %s", err.Error())
			t.FailNow()
		}
		fishies[i] = fish
	}

	if fishies["0"].Type != redfish.Type && fishies["1"].Type != bluefish.Type {
		t.Errorf("Read failed, got %s - %s", fishies, err.Error())
		t.FailNow()
	}
}

//
func TestWriteAndReadEmpty(t *testing.T) {
	createDB(database)
	defer destroySchool()

	// create a fish with no home
	if err := db.Write("", "redfish", redfish); err == nil {
		t.Error("Allowed write of empty collection", err.Error())
		t.FailNow()
	}

	// create a home with no fish
	if err := db.Write(collection, "", redfish); err == nil {
		t.Error("Allowed write of empty resource", err.Error())
		t.FailNow()
	}

	// no place to read
	var onefish Fish
	if err := db.Read("", "redfish", onefish); err == nil {
		t.Error("Allowed read of empty resource", err.Error())
		t.FailNow()
	}

	// no fish to read
	if err := db.Read(collection, "", onefish); err == nil {
		t.Error("Allowed read of empty resource", err.Error())
		t.FailNow()
	}

	// no place to read
	if _, err := db.ReadAll(""); err == nil {
		t.Error("Allowed read of empty resource", err.Error())
		t.FailNow()
	}

	// no place to read
	if _, err := db.ReadAll("nothinghere"); err == nil {
		t.Error("Allowed read of empty resource", err.Error())
		t.FailNow()
	}

	// no place to read
	if _, err := db.ReadAllMap(""); err == nil {
		t.Error("Allowed read of empty resource", err.Error())
		t.FailNow()
	}

	// no place to read
	if _, err := db.ReadAllMap("nothinghere"); err == nil {
		t.Error("Allowed read of empty resource", err.Error())
		t.FailNow()
	}
}

//
func TestDelete(t *testing.T) {
	createDB(database)
	defer destroySchool()

	// add fish to database
	if err := db.Write(collection, "redfish", redfish); err != nil {
		t.Error("Create fish failed: ", err.Error())
		t.FailNow()
	}

	// delete the fish
	if err := db.Delete(collection, "redfish"); err != nil {
		t.Error("Failed to delete: ", err.Error())
		t.FailNow()
	}

	// delete the fish
	if err := db.Delete(collection, "redfish"); err != nil {
		t.Error("Failed to delete: ", err.Error())
		t.FailNow()
	}

	// read fish from database
	var onefish Fish
	if err := db.Read(collection, "redfish", &onefish); err == nil {
		t.Error("Expected nothing, got fish")
		t.FailNow()
	}

	// test delete everything handling
	if err := db.Delete("", "redfish"); err == nil {
		t.Error("Expected nothing, got fish")
		t.FailNow()
	}

}

//
func TestDeleteall(t *testing.T) {
	createDB(database)
	createSchool()
	defer destroySchool()

	if err := destroySchool(); err != nil {
		t.Error("Failed to delete: ", err.Error())
		t.FailNow()
	}

	if _, err := os.Stat(collection); err == nil {
		t.Error("Expected nothing, have fish")
		t.FailNow()
	}
}

// Functions used in testing

// create a new scribble database
func createDB(dir string) error {
	var err error

	if db, err = scribble.New(dir, nil); err != nil {
		return err
	}

	return nil
}

// create a fish
func createFish(fish Fish) error {
	return db.Write(collection, fish.Type, fish)
}

// create many fish
func createSchool() error {
	for i, fish := range []Fish{redfish, bluefish} {
		if err := db.Write(collection, strconv.Itoa(i), fish); err != nil {
			return err
		}
	}

	return nil
}

// destroy all fish
func destroySchool() error {
	return db.Delete(collection, "")
}
