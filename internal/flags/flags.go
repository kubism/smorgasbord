package flags

import (
	"flag"
)

var DexWebDir string

func init() {
	flag.StringVar(&DexWebDir, "dex-web-dir", "", "where to find dex web assets to run tests against")
}
