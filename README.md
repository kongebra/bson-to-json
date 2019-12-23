# BSON to JSON
A simple CLI for converting a BSON-file to a JSON-file

## Usage

### Compile
`go build main.go`

### Flags

```
Usage of bson-to-json
  -d    debug mode
  -o string
        output file (JSON) (default "output.json")
  -p    pretty output
  -s string
        source file (BSON)
```

### Example
```
./bson-to-json -d -p -s /path/to/source.bson -o /path/to/output.json
```
