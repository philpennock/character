package unicode

import (
	"fmt"
	"os"
)

type Unicode struct{}

func Load() Unicode {
	fmt.Fprintf(os.Stderr, "have %d bytes of raw data\n", len(rawData))
	return Unicode{}
}
