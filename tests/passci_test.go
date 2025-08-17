// Package tests mock a test just for passing ci right now
package tests

import (
	"testing"

	"github.com/makestatic/droplink/tests/utils"
)

func TestPass(t *testing.T) {
	// pass
	shift := 1 << 0
	utils.Equals(t, 1, shift)
}
