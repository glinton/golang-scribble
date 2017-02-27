Scribble [![Coverage Status](https://coveralls.io/repos/github/nanobox-io/golang-scribble/badge.svg?branch=master)](https://coveralls.io/github/nanobox-io/golang-scribble?branch=master) [![GoDoc](https://godoc.org/github.com/boltdb/bolt?status.svg)](http://godoc.org/github.com/nanobox-io/golang-scribble) [![Go Report Card](https://goreportcard.com/badge/github.com/nanobox-io/golang-scribble)](https://goreportcard.com/report/github.com/nanobox-io/golang-scribble) [![Build Status](https://travis-ci.org/nanobox-io/golang-scribble.svg)](https://travis-ci.org/nanobox-io/golang-scribble)
--------

A tiny JSON database in Golang


### Installation

Install using `go get github.com/nanobox-io/golang-scribble`.


### Usage

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/nanobox-io/golang-scribble"
)

type Fish struct {
	Type string `json:"type"`
}

var (
	database = "/tmp/school"
	redfish  = Fish{Type: "red"}
)

func main() {
	// a new scribble driver, providing the directory where it will be writing to,
	// and a qualified logger if desired
	db, err := scribble.New(database, nil)
	if err != nil {
		fmt.Println("Error", err)
	}

	// Write a fish to the database
	if err := db.Write("fish", "onefish", redfish); err != nil {
		fmt.Println("Error", err)
	}

	// Read a fish from the database (passing fish by reference)
	fish := Fish{}
	if err := db.Read("fish", "onefish", &fish); err != nil {
		fmt.Println("Error", err)
	}

	// Read all fish from the database, unmarshaling the response.
	records, err := db.ReadAll("fish")
	if err != nil {
		fmt.Println("Error", err)
	}

	var fishies []Fish
	for i := range records {
		fish = Fish{}
		if err = json.Unmarshal([]byte(records[i]), &fish); err != nil {
			fmt.Println("Error", err)
		}
		fishies = append(fishies, fish)
	}

	// Delete a fish from the database
	if err := db.Delete("fish", "onefish"); err != nil {
		fmt.Println("Error", err)
	}

	// Delete all fish from the database
	if err := db.Delete("fish", ""); err != nil {
		fmt.Println("Error", err)
	}
}
```
<!-- See [tests](./scribble_test.go) for more usage. -->

## Documentation

Complete documentation is available on [godoc](http://godoc.org/github.com/nanobox-io/golang-scribble).

## Todo/Doing
- Support for windows
- Better support for concurrency
- Better support for sub collections
- More methods to allow different types of reads/writes
- More tests (you can never have enough!)

## Contributing

Contributions to scribble are welcome and encouraged. Scribble is a [Nanobox](https://nanobox.io) project and contributions should follow the [Nanobox Contribution Process & Guidelines](https://docs.nanobox.io/contributing/).
