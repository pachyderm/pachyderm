package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
)

func Split(buf *bufio.Reader, out chan []byte) error {
	decoder := xml.NewDecoder(buf)

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := token.(type) {
		case xml.StartElement:
			out <- []byte(fmt.Sprintf("+ %s", t.Name.Local))
		case xml.EndElement:
			out <- []byte(fmt.Sprintf("- %s", t.Name.Local))
		default:
		}
	}
}
