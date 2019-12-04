package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	// "github.com/globalsign/mgo/bson"
	// "go.mongodb.org/mongo-driver/bson"
	"github.com/mongodb/mongo-tools/bsondump"
)

var (
	// ErrNotValidJSON error
	ErrNotValidJSON = errors.New("json is not valid")
)

var (
	input  string
	output string
	pretty bool
	check  bool
	debug  bool
)

func init() {
	flag.StringVar(&input, "input", "", "(Required) Input file")
	flag.StringVar(&input, "i", "", "(Required) Input file (shorthand)")

	flag.StringVar(&output, "output", "data.json", "Output file default: 'data.json')")
	flag.StringVar(&output, "o", "data.json", "Output file default: 'data.json') (shorthand)")

	flag.BoolVar(&pretty, "pretty", false, "Pretty output (default: false)")
	flag.BoolVar(&pretty, "p", false, "Pretty output (default: false) (shorthand)")

	flag.BoolVar(&check, "check", false, "Check and validate BSON during processing (default: false)")
	flag.BoolVar(&check, "c", false, "Check and validate BSON during processing (default: false) (shorthand)")

	flag.BoolVar(&debug, "debug", false, "debug")
	flag.BoolVar(&debug, "d", false, "debug")
}

func main() {
	defer measure(time.Now(), "main")

	// Parse flags
	flag.Parse()
	checkFlags()

	err := BSON2JSON()
	if err != nil {
		log.Fatal(err)
	}

	data, err := getOutputData()
	if err != nil {
		log.Fatal(err)
	}

	valid := validateOutputData(data)
	if !valid {
		err := fix(data)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkFlags() {
	// Check if input is blank
	if input == "" {
		log.Fatal("flag needs an argument: -input")
	}

	// Check if input is a BSON document
	if !strings.HasSuffix(input, ".bson") {
		log.Fatal("file type error: input file needs to be a BSON document")
	}

	// Check if output is a JSON file
	if !strings.HasSuffix(output, ".json") {
		log.Fatal("file type error: output file ened to be a JSON file")
	}
}

func measure(start time.Time, text string) {
	if debug {
		fmt.Printf("%s took %s\n", text, time.Since(start))
	}
}

// BSON2JSON func
func BSON2JSON() error {
	defer measure(time.Now(), "converting BSON to JSON")

	// BSON dump options
	options := bsondump.Options{
		ToolOptions: nil,
		OutputOptions: &bsondump.OutputOptions{
			Type:         bsondump.JSONOutputType,
			ObjCheck:     check,
			Pretty:       pretty || true,
			BSONFileName: input,
			OutFileName:  output,
		},
	}

	// Create BSON dump
	bd, err := bsondump.New(options)
	if err != nil {
		return err
	}
	// Close up, eventually
	defer bd.Close()

	// Create new output file
	file, err := os.Create(output)
	if err != nil {
		return err
	}

	// Close output file, eventually
	defer file.Close()

	// Set output file as output writer
	bd.OutputWriter = file

	// Convert to JSON
	_, err = bd.JSON()
	if err != nil {
		return err
	}

	return nil
}

func getOutputData() ([]byte, error) {
	return ioutil.ReadFile(output)
}

func validateOutputData(data []byte) bool {
	defer measure(time.Now(), "validate output JSON")

	// Check if new data is valid
	return json.Valid(data)
}

func fix(data []byte) error {
	defer measure(time.Now(), "fix JSON")

	bytesReader := bytes.NewReader(data)

	// Set first to true
	first := true

	// Create a new scanner
	scanner := bufio.NewScanner(bytesReader)

	var sb strings.Builder
	// Append a bracket on the start
	sb.WriteString("[\n")

	// Scan lines
	for scanner.Scan() {
		// Get text in line
		line := scanner.Text()

		// TODO (Svein): This does not work with non-pretty
		// Check if line is equal to '{' and not the first line
		if line == "{" && !first {
			sb.WriteString(",\n")
		}

		// Write line
		sb.WriteString(strings.ReplaceAll(line, "\t", ""))

		// Set first line to false
		if first {
			first = false
		}
	}

	// Write last line
	sb.WriteString("\n]")

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return err
	}

	// Check if new data is valid
	valid := json.Valid([]byte(sb.String()))
	if !valid {
		return ErrNotValidJSON
	}

	// Write to file
	return ioutil.WriteFile(output, []byte(sb.String()), 0644)
}
