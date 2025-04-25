//go:build ignore

package main

// This is a simple test program which can be run like so:
//     go run otest.go < a.ogg > b.ogg
// It's quite simple in that it doesn't handle recombining COP packets,
// but at least for the ogg/vorbis files from wikipedia that I've tested,
// it results in an identical copy of the original file.

import (
	"fmt"
	"os"

	"github.com/SaurusXI/ogg"
)

func main() {
	decoder := ogg.NewDecoder(os.Stdin)
	page, _, err := decoder.Decode()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	encoder := ogg.NewEncoder(page.Serial, os.Stdout)
	encoder.EncodeBOS(page.Granule, page.Packets)

	for {
		page, _, err := decoder.Decode()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			break
		}

		if page.Type&ogg.EOS == ogg.EOS {
			encoder.EncodeEOS(page.Granule, page.Packets)
			break
		}
		encoder.Encode(page.Granule, page.Packets)
	}
}
