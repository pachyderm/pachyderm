package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func SplitJson(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(strings.TrimSpace(string(data))) == 0 {
		return 0, nil, nil // No more records; finish scanning
	}
	if i := bytes.IndexByte(data, '}'); i > 0 {
		return i + 1, data[:i+1], nil // Encountered finished json record
	} else if atEOF {
		return 0, nil, fmt.Errorf("incomplete JSON record: file did not end in '}'")
	}
	return 0, nil, nil // request more data
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: to_sql takes one argument (name of the "+
			"table receiving writes), but received none\n")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(SplitJson)
	for scanner.Scan() {
		// Parse json as plain interface{}. See https://blog.golang.org/json-and-go
		var jsinter interface{}
		json.Unmarshal(scanner.Bytes(), &jsinter)
		js, found := jsinter.(map[string]interface{})
		if !found {
			log.Fatalf("Could not parse json (may be nested):\n\"\"\"\n%s\n\"\"\"",
				scanner.Text())
		}

		// Convert json into SQL INSERT
		keys := make([]string, len(js))
		values := make([]string, len(js))
		i := 0
		for k, v := range js {
			keys[i] = k
			values[i] = fmt.Sprintf("%v", v)
			_, ismap := v.(map[string]interface{})
			_, isinterface := v.([]interface{})
			if ismap || isinterface {
				log.Fatalf("Could not convert nested json value: \"%v\"", values[i])
			}
			i++
		}
		fmt.Printf("INSERT INTO %s (%s) VALUES (%s);\n",
			os.Args[1],
			strings.Join(keys, ", "),
			strings.Join(values, ", "))
	}
	if scanner.Err() != nil {
		log.Fatalf("Could not scan json record: %s", scanner.Err().Error())
	}
}
