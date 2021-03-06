// Copyright (c) Jeevanandam M (https://github.com/jeevatkm)
// Source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package ainsp

import (
	"os"
	"path/filepath"
	"testing"

	"aahframe.work/essentials"
	"github.com/stretchr/testify/assert"
)

func TestProgramRandomLoad(t *testing.T) {
	gopath := os.Getenv("GOROOT")
	prgPath := filepath.Join(gopath, "src", "net", "http", "httputil")

	prg, err := Inspect(prgPath, "net/http/httputil", ess.Excludes([]string{}), map[string]map[string]uint8{})
	assert.NotNil(t, err)
	assert.NotNil(t, prg)
}
