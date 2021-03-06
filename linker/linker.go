package linker

import (
	"errors"
	"strings"

	parser "tritium/parser"
	tp "tritium/proto"
)

func RunStringWithPackage(src, projectPath, scriptPath, fileName string, pkg *tp.Package, layers []string) (*tp.Transform, error) {
	objs := parser.Parse(src, projectPath, scriptPath, fileName, false, layers)
	return runWithObjs(objs, pkg, projectPath, scriptPath, layers)
}

func RunWithPackage(projectPath, scriptPath, fileName string, pkg *tp.Package, layers []string) (*tp.Transform, error) {
	objs := parser.ParseFileSet(projectPath, scriptPath, fileName, false, layers)
	return runWithObjs(objs, pkg, projectPath, scriptPath, layers)
}

func runWithObjs(objs []*tp.ScriptObject, pkg *tp.Package, projectPath, scriptPath string, layers []string) (*tp.Transform, error) {
	ctx := NewObjectLinkingContext(pkg, objs, projectPath, scriptPath)
	ctx.Link()
	if ctx.HasErrors() {
		message := ""
		for _, msg := range ctx.Errors {
			message = message + "\n" + msg
		}
		return nil, errors.New(message)
	}

	ls := strings.Join(layers, "/")
	ctx.Transform.Layers = &ls
	return ctx.Transform, nil
}
