/*
Open Source Initiative OSI - The MIT License (MIT):Licensing
The MIT License (MIT)
Copyright (c) 2017 Ralph Caraveo (deckarep@gmail.com)
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
)

// Usage: go run main.go --help (for help)
// Usage: go run main.go -k "chef_environment:production* AND role:filter"  -m 'sudo which jq || sudo yum -y install jq' -C 5
const (
	queryBucket = "queries"
)

// TODO: move ssh crap into it's own package
var db *bolt.DB

// Idea make this a running agent?
// Agent is always refreshing in the background your most common queries?
// It can update every 5 minutes and on failures will just log the error.

func storeBolt(key, value string) {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(queryBucket))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), []byte(value))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func getBolt(key string) string {
	result := ""
	// retrieve the data
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(queryBucket))
		if bucket == nil {
			// Bucket not created yet
			return nil
		}

		val := bucket.Get([]byte(key))
		result = string(val)

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func openDB() {
	var err error
	db, err = bolt.Open("blade-boltdb.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func endOnSignal() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	<-done
}
