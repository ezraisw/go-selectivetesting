package app

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var outMut sync.Mutex

type multiError []error

func (errs multiError) Error() string {
	msgs := make([]string, 0, len(errs))
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

func runTests(moduleDir, args string, parallel int, testedPkgs []testedPackage) error {
	if parallel < 1 {
		parallel = 1
	}

	var (
		runErrsMut sync.Mutex
		runErrs    multiError
	)

	wg := sync.WaitGroup{}
	qc := make(chan struct{}, parallel)
	for _, testedPkg := range testedPkgs {
		wg.Add(1)
		qc <- struct{}{}

		go func(testedPkg testedPackage) {
			defer func() {
				<-qc
				wg.Done()
			}()

			cmd := exec.Command("go", "test", testedPkg.PkgPath, "-run", testedPkg.RunRegex, args)
			cmd.Dir = moduleDir

			stderrBuf := &bytes.Buffer{}
			cmd.Stderr = stderrBuf

			stdoutBuf := &bytes.Buffer{}
			cmd.Stdout = stdoutBuf

			if err := cmd.Run(); err != nil {
				runErrsMut.Lock()
				runErrs = append(runErrs, err)
				runErrsMut.Unlock()

				switch err.(type) {
				case *exec.ExitError:
					break // Do not return and let it print.
				case *exec.Error:
					return
				default:
					return
				}
			}

			outMut.Lock()
			defer outMut.Unlock()

			fmt.Fprintln(os.Stdout, "cmd:", cmd.String())
			_, _ = io.Copy(os.Stderr, stderrBuf)
			_, _ = io.Copy(os.Stdout, stdoutBuf)
		}(testedPkg)
	}
	wg.Wait()

	if len(runErrs) > 0 {
		return runErrs
	}
	return nil
}
