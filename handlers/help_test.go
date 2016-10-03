package handlers

import (
	"strconv"
	"testing"
)

func TestHelp(t *testing.T) {
	message := help()
	if len(message) != 126 {
		t.Error("help message was this long: " + strconv.Itoa(len(message)))
	}
}
