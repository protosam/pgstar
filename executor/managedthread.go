package executor

import (
	"os"
	"path/filepath"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type ManagedThread struct {
	*starlark.Thread
	starfile     string
	rootdir      string
	predeclared  starlark.StringDict
	moduleLoader *ModuleLoader
}

var pwd string

func NewManagedThread(rootdir, starfile string) *ManagedThread {
	// store the absolute path to the rootdir
	rootdir, _ = filepath.Abs(rootdir)

	// relative path to the script from the working directory
	pwd, _ = os.Getwd()
	name, _ := filepath.Rel(pwd, filepath.Join(rootdir, starfile))

	// populate the ManagedThread struct
	mt := &ManagedThread{
		Thread: &starlark.Thread{
			Name:  name,
			Print: PrintLog,
		},
		predeclared: starlark.StringDict{},
		starfile:    starfile,
		rootdir:     rootdir,
	}

	// mt.moduleLoader = NewModuleLoader(mt, rootdir, starfile)
	return mt
}

func (mt *ManagedThread) NewChild(starfile string) *ManagedThread {
	name, _ := filepath.Rel(pwd, filepath.Join(mt.rootdir, starfile))
	childLoader := mt.moduleLoader.NewChild(starfile)
	return &ManagedThread{
		Thread: &starlark.Thread{
			Name:  name,
			Print: mt.Thread.Print,
			Load:  childLoader.Load,
		},
		starfile:     starfile,
		rootdir:      mt.rootdir,
		predeclared:  mt.predeclared,
		moduleLoader: childLoader,
	}
}

func (mt *ManagedThread) GetStarfile() string {
	return mt.starfile
}

func (mt *ManagedThread) GetRootdir() string {
	return mt.rootdir
}

func (mt *ManagedThread) SetModuleLoader(loader *ModuleLoader) {
	mt.moduleLoader = loader
	mt.Thread.Load = loader.Load
}

func (mt *ManagedThread) Predeclare(name string, value starlark.Value) {
	mt.predeclared[name] = value
}

func (mt *ManagedThread) Exec() (starlark.StringDict, error) {
	fileOptions := &syntax.FileOptions{
		GlobalReassign:  true,
		TopLevelControl: true,
	}
	return starlark.ExecFileOptions(fileOptions, mt.Thread, mt.Name, nil, mt.predeclared)
}
