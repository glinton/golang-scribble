// Package scribble is a tiny JSON database
package scribble

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jcelliott/lumber"
)

type (

	// Logger is a generic logger interface
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	// Driver is what is used to interact with the scribble database. It runs
	// transactions, and provides log output
	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]sync.Mutex
		dir     string // the directory where scribble will create the database
		log     Logger // the logger scribble will log to
	}
)

// Options uses for specification of working golang-scribble
type Options struct {
	Logger // the logger scribble will use (configurable)
}

// New creates a new scribble database at the desired directory location, and
// returns a *Driver to then use for interacting with the database
func New(dir string, options *Options) (*Driver, error) {

	// ensure root is not specified (delete could wipe their drive/files) (filepath.Clean won't end in a '/')
	dir = filepath.Clean(dir)
	if dir == "/" || dir == "C:\\" || dir == "~" || (strings.HasPrefix(dir, "/home/") && strings.Count(dir, "/") <= 2) {
		return nil, fmt.Errorf("Missing or unsafe filepath - no place to create db!")
	}

	// create default options
	opts := Options{}

	// if options are passed in, use those
	if options != nil {
		opts = *options
	}

	// if no logger is provided, create a default
	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	//
	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]sync.Mutex),
		log:     opts.Logger,
	}

	// if the database already exists, just use it
	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &driver, nil
	}

	// if the database doesn't exist create it
	opts.Logger.Debug("Creating scribble database at '%s'...\n", dir)
	return &driver, os.MkdirAll(dir, 0755)
}

// Write locks the database and attempts to write the record to the database under
// the [collection] specified with the [resource] name given
func (d *Driver) Write(collection, resource string, v interface{}) error {

	// ensure there is a place to save record
	if collection == "" {
		return fmt.Errorf("Missing collection - no place to save record!")
	}

	// ensure there is a resource (name) to save record as
	if resource == "" {
		return fmt.Errorf("Missing resource - unable to save record - no name!")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	//
	dir := filepath.Join(d.dir, collection)
	fnlPath := filepath.Join(dir, resource)
	tmpPath := fnlPath + ".tmp"

	// create collection directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	//
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	// write marshaled data to the temp file
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	// move final file into place
	return os.Rename(tmpPath, fnlPath)
}

// Read a record from the database
func (d *Driver) Read(collection, resource string, v interface{}) error {

	// ensure there is a place to read record
	if collection == "" {
		return fmt.Errorf("Missing collection - no place to read record!")
	}

	// ensure there is a resource (name) to read record from
	if resource == "" {
		return fmt.Errorf("Missing resource - unable to read record - no name!")
	}

	//
	record := filepath.Join(d.dir, collection, resource)

	// check to see if file exists
	if _, err := stat(record); err != nil {
		return err
	}

	// read record from database
	b, err := ioutil.ReadFile(record)
	if err != nil {
		return err
	}

	// unmarshal data
	return json.Unmarshal(b, &v)
}

// ReadAll records from a collection; this is returned as a slice of strings because
// there is no way of knowing what type the record is.
func (d *Driver) ReadAll(collection string) ([]string, error) {

	// ensure there is a collection to read
	if collection == "" {
		return nil, fmt.Errorf("Missing collection - unable to read location!")
	}

	//
	dir := filepath.Join(d.dir, collection)

	// check to see if collection (directory) exists
	if _, err := stat(dir); err != nil {
		return nil, fmt.Errorf("Directory '%s' does not exist - %s!", dir, err.Error())
	}

	// read all the files in the transaction.Collection; an error here just means
	// the collection is either empty or doesn't exist
	files, _ := ioutil.ReadDir(dir)

	// the files read from the database
	var records []string

	// iterate over each of the files, attempting to read the file. If successful
	// append the files to the collection of read files
	for i := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, files[i].Name()))
		if err != nil {
			return nil, err
		}

		// append read file
		records = append(records, string(b))
	}

	// unmarhsal the read files as a comma delimeted byte array
	return records, nil
}

// ReadAllMap records from a collection; this is returned as a string map of strings
// because the resource was a string, and there is no way of knowing what type
// the record is.
func (d *Driver) ReadAllMap(collection string) (map[string]string, error) {

	// ensure there is a collection to read
	if collection == "" {
		return nil, fmt.Errorf("Missing collection - unable to read location!")
	}

	//
	dir := filepath.Join(d.dir, collection)

	// check to see if collection (directory) exists
	if _, err := stat(dir); err != nil {
		return nil, fmt.Errorf("Directory '%s' does not exist - %s!", dir, err.Error())
	}

	// read all the files in the transaction.Collection; an error here just means
	// the collection is either empty or doesn't exist
	files, _ := ioutil.ReadDir(dir)

	// the files read from the database (map[string] because the resource is a string)
	var records = make(map[string]string)

	// iterate over each of the files, attempting to read the file. If successful
	// append the files to the collection of read files
	for i := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, files[i].Name()))
		if err != nil {
			return nil, err
		}

		// append read file
		// records = append(records, string(b))
		records[files[i].Name()] = string(b)
	}

	// unmarhsal the read files as a comma delimeted byte array
	return records, nil
}

// Delete locks that database and then attempts to remove the collection/resource
// specified by [path]
func (d *Driver) Delete(collection, resource string) error {

	// ensure there is a place to delete record. it's fine if a resource is blank,
	// this would be a deleteAll equivalent
	resource = filepath.Clean(resource)
	collection = filepath.Clean(collection)
	if collection == "" || collection == "/" || collection == "." {
		return fmt.Errorf("Missing collection - no place to delete record!")
	}

	path := filepath.Join(collection, resource)
	//
	mutex := d.getOrCreateMutex(path)
	mutex.Lock()
	defer mutex.Unlock()

	//
	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir); {

	// if fi is nil or error is not nil return
	case fi == nil, err != nil:
		if strings.Contains(err.Error(), "no such file") {
			return nil
		}
		return fmt.Errorf("Unable to stat %s - %s!", dir, err.Error())

	// remove file or directory and all contents
	case fi.Mode().IsDir(), fi.Mode().IsRegular():
		return os.RemoveAll(dir)
	}

	return nil
}

// stat checks for dir, if path isn't a directory check to see if it's a file
func stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// getOrCreateMutex creates a new collection specific mutex any time a collection
// is being modfied to avoid unsafe operations
func (d *Driver) getOrCreateMutex(collection string) sync.Mutex {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	m, ok := d.mutexes[collection]

	// if the mutex doesn't exist make it
	if !ok {
		m = sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}
