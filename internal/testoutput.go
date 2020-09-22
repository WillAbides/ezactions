package testoutput

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/willabides/ezactions"
)

type testEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

func (e *testEvent) key() string {
	return strings.Join([]string{e.Package, e.Test}, ":")
}

func (te testEvents) ByAction() testEventsMaps {
	m := testEventsMaps{}
	for _, ev := range te {
		m[ev.Action] = append(m[ev.Action], ev)
	}
	return m
}

func (te testEvents) byKey() testEventsMaps {
	m := testEventsMaps{}
	for _, ev := range te {
		m[ev.key()] = append(m[ev.key()], ev)
	}
	return m
}

func (te testEvents) withTest() testEvents {
	res := testEvents{}
	for _, event := range te {
		if event.Test != "" {
			res = append(res, event)
		}
	}
	return res
}

func (te testEvents) withPackage() testEvents {
	res := testEvents{}
	for _, event := range te {
		if event.Package != "" && event.Package != "command-line-arguments" {
			res = append(res, event)
		}
	}
	return res
}

func (te testEvents) sortEvents() testEvents {
	sort.Slice(te, func(i, j int) bool {
		return te[i].Time.Before(te[j].Time)
	})
	return te
}

func (te testEvents) output() string {
	output := ""
	for _, ev := range te.ByAction()["output"].sortEvents() {
		output += ev.Output
	}
	return output
}

func (te testEvents) result() *testEvent {
	for _, event := range te {
		for _, res := range resultActions {
			if event.Action == res {
				return event
			}
		}
	}
	return nil
}

type testEventsMaps map[string]testEvents

func (m testEventsMaps) sortedKeys() []string {
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (m testEventsMaps) filterByResult(desiredResult string) testEventsMaps {
	out := make(testEventsMaps)
	for key, events := range m {
		event := events.result()
		if event == nil {
			continue
		}
		result := event.Action
		if result == desiredResult {
			out[key] = events
		}
	}
	return out
}

func parseEvents(reader io.Reader, passthrough io.Writer) testEvents {
	events := testEvents{}
	jsonScanner := bufio.NewScanner(reader)
	for jsonScanner.Scan() {
		event := new(testEvent)
		err := json.Unmarshal(jsonScanner.Bytes(), event)
		if err != nil {
			continue
		}
		if passthrough != nil && event.Action == "output" {
			_, err := passthrough.Write([]byte(event.Output))
			if err != nil {
				panic(err)
			}
		}

		events = append(events, event)
	}

	return events
}

var resultActions = []string{"pass", "fail"}

type testEvents []*testEvent

func OutputFailures(input io.Reader, output io.Writer, rootPath, rootPkg string, passthrough bool) int {
	commander := &ezactions.WorkflowCommander{
		Printer: func(s string) {
			fmt.Fprint(output, s)
		},
	}
	var passthroughWriter io.Writer
	if passthrough {
		passthroughWriter = output
	}
	events := parseEvents(input, passthroughWriter)
	failingTests := events.withTest().withPackage().byKey().filterByResult("fail")
	for _, key := range failingTests.sortedKeys() {
		events := failingTests[key]
		resEvent := events.result()
		if resEvent == nil {
			continue
		}
		pkg := resEvent.Package
		testName := resEvent.Test
		var loc *ezactions.CommanderFileLocation
		testFile, testLine, err := findTest(pkg, testName, rootPath, rootPkg)
		if err == nil && testLine != 0 {
			loc = &ezactions.CommanderFileLocation{
				File: testFile,
				Line: testLine,
			}
		}
		commander.SetErrorMessage(resEvent.Output, loc)
	}

}

func findTest(pkg, testName, rootPath, rootPkg string) (string, int, error) {
	if !strings.HasPrefix(pkg, rootPkg) {
		return "", 0, fmt.Errorf("%s does not contain %s", rootPkg, pkg)
	}
	relPkg := strings.TrimPrefix(pkg, rootPkg)
	relPkg = filepath.FromSlash(relPkg)
	dir := filepath.Join(rootPath, relPkg)
	dirstat, err := os.Stat(dir)
	if err != nil {
		return "", 0, errors.New("failed statting directory: " + err.Error())
	}
	if !dirstat.IsDir() {
		return "", 0, fmt.Errorf("not a directory: %q", dir)
	}
	testName = strings.Split(testName, "/")[0]
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return "", 0, errors.New("failed parsing directory: " + err.Error())
	}
	var testFile string
	var testLine int
	for _, pkg := range pkgs {
		ast.Inspect(pkg, func(n ast.Node) bool {
			decl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if decl.Name.Name == testName {
				p := fset.Position(decl.Pos())
				testFile = p.Filename
				testLine = p.Line
			}
			return true
		})
		if testFile != "" {
			break
		}
	}
	testFile, err = filepath.Rel(rootPath, testFile)
	if err != nil {
		return "", 0, err
	}
	testFile = strings.TrimPrefix(filepath.Join("..", testFile), ".")
	return testFile, testLine, nil
}