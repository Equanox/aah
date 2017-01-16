// Copyright (c) Jeevanandam M (https://github.com/jeevatkm)
// go-aah/aah source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package aruntime

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"aahframework.org/config"
	"aahframework.org/essentials"
)

const (
	goroutinePrefix = "goroutine"
	createdByPrefix = "created by"
)

var goroutineRegEx = regexp.MustCompile(`goroutine\s?\d+\s\[.*\]\:`)

type (
	// Stacktrace holds the parse information of `debug.Stack()`. It's easier to
	// debug and understand.
	Stacktrace struct {
		Raw        string
		Recover    interface{}
		RoutineCnt int
		IsParsed   bool
		GoRoutines []*GoRoutine

		maxFileLen int
		gopathSrc  string
		gorootSrc  string
	}

	// GoRoutine holds information of single Go routine stack trace.
	GoRoutine struct {
		Header    string
		Packages  []string
		Functions []string
	}
)

// NewStacktrace method collects debug stack information and parsing them into
// easy understanding and returns the instance.
func NewStacktrace(r interface{}, appCfg *config.Config) *Stacktrace {
	strace := &Stacktrace{
		Raw:     string(debug.Stack()),
		Recover: r,
	}

	if appCfg.BoolDefault("runtime.all_goroutines", false) {
		buf := make([]byte, 2<<20) // TODO implement config size instead of hardcode 2mb
		length := runtime.Stack(buf, true)
		if length < len(buf) {
			buf = buf[:length]
		}

		strace.Raw = string(buf)
	} else {
		strace.Raw = string(debug.Stack())
	}

	strace.initPath()

	return strace
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Stacktrace methods
//___________________________________

// Parse method parses the go debug stacktrace into easy to understand.
func (st *Stacktrace) Parse() {
	routines := goroutineRegEx.FindAllString(st.Raw, -1)
	st.RoutineCnt = len(routines)
	st.GoRoutines = make([]*GoRoutine, st.RoutineCnt)

	ri := -1
	lines := strings.Split(st.Raw, "\n")
	gopathSrcLen := len(st.gopathSrc) + 1
	gorootSrcLen := len(st.gorootSrc) + 1

	for linePos := 0; linePos < len(lines); linePos++ {
		sline := strings.TrimSpace(lines[linePos])
		if len(sline) == 0 {
			continue
		}

		if strings.HasPrefix(sline, goroutinePrefix) {
			ri++
			st.GoRoutines[ri] = &GoRoutine{
				Header:    sline,
				Packages:  []string{},
				Functions: []string{},
			}

			continue
		}

		if strings.HasPrefix(sline, "/") {
			if strings.HasPrefix(sline, st.gopathSrc) {
				sline = sline[gopathSrcLen:]
			} else if strings.HasPrefix(sline, st.gorootSrc) {
				sline = sline[gorootSrcLen:]
			}

			sline = sline[:strings.LastIndex(sline, " ")]
			if len(sline) > st.maxFileLen {
				st.maxFileLen = len(sline)
			}

			st.GoRoutines[ri].Packages = append(st.GoRoutines[ri].Packages, sline)
		} else {
			isCreatedBy := strings.HasPrefix(sline, createdByPrefix)
			sline = filepath.Base(sline)

			if !isCreatedBy {
				rparen := strings.LastIndex(sline, "(")
				comma := strings.IndexByte(sline[rparen:], ',')
				if comma == -1 {
					sline = sline[:rparen+1] + ")"
				} else {
					sline = sline[:rparen+1] + " ... )"
				}
			}

			st.GoRoutines[ri].Functions = append(st.GoRoutines[ri].Functions, sline)
		}

	}

	st.IsParsed = true
}

// Print method prints the stack trace info to io.Writer.
func (st *Stacktrace) Print(w io.Writer) {
	if w == nil {
		return
	}

	if !st.IsParsed {
		st.Parse()
	}

	printFmt := "\t%-" + strconv.Itoa(st.maxFileLen+1) + "s-> %v\n"
	_, _ = w.Write([]byte(fmt.Sprintf("\n%v\n", st.Recover)))

	for _, rv := range st.GoRoutines {
		_, _ = w.Write([]byte("\n" + rv.Header + "\n"))
		for idx, f := range rv.Packages {
			_, _ = w.Write([]byte(fmt.Sprintf(printFmt, f, rv.Functions[idx])))
		}
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Unexported methods
//___________________________________

func (st *Stacktrace) initPath() {
	gopath, _ := ess.GoPath()
	goroot := runtime.GOROOT()

	st.gopathSrc = filepath.Join(gopath, "src")
	st.gorootSrc = filepath.Join(goroot, "src")
}
