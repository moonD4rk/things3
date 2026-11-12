package cmd

import (
	"os"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

// TestMain pins every command's client to the fixture clock, so the fixture's
// dates keep their meaning (upcoming stays upcoming) as real time moves on.
func TestMain(m *testing.M) {
	clock = thingstest.Now
	os.Exit(m.Run())
}
