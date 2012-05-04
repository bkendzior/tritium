package spec

import (
	tp "athena"
	. "fmt"
	"fmt"
	xmlhelp "gokogiri/help"
	l4g "log4go"
	"os"
	. "path/filepath"
	"runtime/debug"
	. "tritium"
	"tritium/lamprey"
	"tritium/packager"
	"tritium/whale"
)

func All(command string, directory string, options ...string) {

	var mixerPath string
	if len(options) == 1 {
		mixerPath = options[0]
	}

	logger := make(l4g.Logger)
	logger.AddFilter("test", l4g.ERROR, l4g.NewConsoleLogWriter())
	l4g.Global = logger
	var eng Engine
	if command == "test" {
		eng = whale.NewEngine(logger)
	} else if command == "debug" {
		eng = lamprey.NewEngine(logger)
	}

	var pkg *tp.Package

	if len(mixerPath) > 0 {
		// Used when testing in ambrosia
		mixer := tp.OpenMixer(mixerPath)
		pkg = mixer.Package
	} else {
		bigPackage := packager.BuildDefaultPackage()
		pkg = bigPackage.Package
	}

	globalResult := NewResult()
	globalResult.all(directory, pkg, eng, logger)

	logger.AddFilter("stdout", l4g.ERROR, l4g.NewConsoleLogWriter())

	// TODO : Walk over the results here and print errors. 

	var foundError = false

	for _, err := range globalResult.Errors {
		foundError = true
		println("\n=========================================", err.Location, "\n")
		if err.Panic {
			Printf(err.Message)
		} else {
			Printf("\n==========\n%v :: %v \n\n Got \n----------\n%v\n\n Expected \n----------\n%v\n", err.Name, err.Message, err.Got, err.Expected)
		}
	}
	println("\n\n")
	println("+++TEST COMPLETE+++\n\n")

	if foundError {
		os.Exit(1)
	}
	eng.Free()
	xmlhelp.LibxmlCleanUpParser()
	if xmlhelp.LibxmlGetMemoryAllocation() != 0 {
		fmt.Printf("Memeory leaks %d!!!", xmlhelp.LibxmlGetMemoryAllocation())
		xmlhelp.LibxmlReportMemoryLeak()
	}
}

func (result *Result) all(directory string, pkg *tp.Package, eng Engine, logger l4g.Logger) {
	paths, err := Glob(Join(directory, "main.ts"))
	if err == nil && len(paths) == 1 {
		newResult := RunSpec(directory, pkg, eng, logger)
		result.Merge(newResult)
	}

	subdirs, _ := Glob(Join(directory, "*"))
	for _, subdir := range subdirs {
		fi, err := os.Stat(subdir)
		if err != nil {
			continue
		}
		if !fi.IsDir() {
			continue
		}
		result.all(subdir, pkg, eng, logger)
	}

}

func RunSpec(dir string, pkg *tp.Package, eng Engine, logger l4g.Logger) (result *Result) {
	result = NewResult()
	logWriter := NewTestLogWriter()
	logger["test"] = &l4g.Filter{l4g.WARNING, "test", logWriter}

	defer func() {
		//log.Println("done")  // Println executes normally even in there is a panic
		if x := recover(); x != nil {
			err, ok := x.(error)
			if ok {
				logger.Error(dir + " === " + err.Error() + "\n\n" + string(debug.Stack()))
			} else {
				logger.Error(dir + " === " + x.(string) + "\n\n" + string(debug.Stack()))
			}
		}
		for _, rec := range logWriter.Logs {
			//println("HAZ LOGS")
			err := l4g.FormatLogRecord("[%D %T] [%L] (%S) %M", rec)
			result.Error(dir, err)
		}
		print(result.CharStatus())
	}()

	spec, err := LoadSpec(dir, pkg)
	if err != nil {
		result.Error(dir, err.Error())
	} else {
		result.Merge(spec.Compare(eng.Run(spec.Script, spec.Input, spec.Vars)))
	}
	return
}
