/**
* All common util functions and golbal constants will go here.
**/
package acceptance

import (
	"os"
	"time"
)

const EVENTUALLY_DEFAULT_TIME_OUT time.Duration = 60 * time.Second

var WEGO_BIN_PATH string

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
