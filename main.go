package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
)

var (
	sourceFile string
	outputFile string
	pretty     bool
	debug      bool
)

func init() {
	flag.StringVar(&sourceFile, "s", "", "source file (BSON)")
	flag.StringVar(&outputFile, "o", "output.json", "output file (JSON)")
	flag.BoolVar(&pretty, "p", false, "pretty output")
	flag.BoolVar(&debug, "d", false, "debug mode")

	flag.Parse()
}

func writeFile(v interface{}) {
	var data []byte
	var err error

	// Check if we want pretty print
	if pretty {
		data, err = json.MarshalIndent(v, "", "\t")
	} else {
		data, err = json.Marshal(v)
	}

	// Look for errors
	if err != nil {
		log.Fatal(err)
	}

	// Write file
	err = ioutil.WriteFile(outputFile, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Measure time if we want debugging
	if debug {
		defer measure(
			time.Now(),
			fmt.Sprintf("Converting \"%s\" to JSON", sourceFile),
		)
	}

	// Declare some variables
	var (
		offset  int32
		docSize int32
		reader  io.Reader
		buffer  *bytes.Buffer
	)

	// Result slice of BSON objects
	result := make([]bson.M, 0)

	// Read data from file
	data, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		panic(err)
	}

	// Run forever
	for {
		// Create new bytes reader based on offset
		reader = bytes.NewReader(data[offset:])

		// Read binary to bytes reader
		if err := binary.Read(reader, binary.LittleEndian, &docSize); err != nil {
			// Check if we have reached end of file
			if err == io.EOF {
				break
			}

			// Panic, get out!
			panic(err)
		}

		// Min and max bson document sizes allowed
		if docSize < 5 || docSize > 16777216 {
			// Panic with message
			panic("invalid document size")
		}

		// Create new bytes buffer with the capacity of the document size
		buffer = bytes.NewBuffer(make([]byte, 0, docSize))

		// Write binaries to the bytes buffer
		if err := binary.Write(buffer, binary.LittleEndian, docSize); err != nil {
			// Panic on the dance floor!
			panic(err)
		}

		// Copy N size from reader to buffer
		if _, err := io.CopyN(buffer, reader, int64(docSize-4)); err != nil {
			// Panic at the Disco
			panic(err)
		}

		// Declare document
		var document bson.M

		// Unmarshal bytes to an interface
		if err := bson.Unmarshal(buffer.Bytes(), &document); err != nil {
			// P for panic
			panic(err)
		}

		// Append document to result
		result = append(result, document)

		// Add document size to offset
		offset += docSize
	}

	// Write result to file
	writeFile(result)
}

func measure(start time.Time, text string) {
	fmt.Printf("%s took %s\n", text, time.Since(start))
}
