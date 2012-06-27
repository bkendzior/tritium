package packager

// TODO(SJ): add hooks to check features here
import (
	tp "tritium/proto"
	proto "code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	yaml "goyaml"
	"io/ioutil"
	"path/filepath"
	"strings"
	"os"
)

type Mixer struct {
	*tp.Mixer
	LibraryPath    string
	DefinitionPath string
	DataPath       string
	RootPackage    *Package // I need to keep the tritium Package around for its methods until I'm done resolving it
}

var SubPaths [3]string

func init() {
	SubPaths = [3]string{"external", "internal", ""}
}

func GetDataPath() (path string, err error) {
	path = os.ExpandEnv("$GOHATTAN_DATA")

	if path == "" {
		defaultPath := os.ExpandEnv("$HOME/.manhattan")
		_, err := ioutil.ReadDir(defaultPath)

		if err != nil {
			return "", errors.New("Data path ($GOHATTAN_DATA) not set and default path ~/.manhattan doesn't exist.")
		} else {
			path = defaultPath
		}
	}

	return
}

func BuildMixer(buildPath string, name string, dataPath string) *Mixer {
	rawPath, err := findInBuildPath(buildPath, name)

	if err != nil {
		panic(err.Error())
	}

	path := *rawPath

	raw_mixer := tp.NewMixer(filepath.Clean(path))
	mixer := &Mixer{
		Mixer:          raw_mixer,
		LibraryPath:    buildPath,
		DefinitionPath: path,
		DataPath:       dataPath,
	}

	rewritersDirectory := filepath.Join(path, "/rewriters")
	mixer.Rewriters = tp.CollectFiles(rewritersDirectory)

	packageDirectory := filepath.Join(path, "/package")
	mixer.RootPackage = NewRootPackage(packageDirectory, proto.GetString(mixer.Name), dataPath)

	mixer.loadDependentMixers()

	error := BuildRootPackage(mixer.RootPackage, packageDirectory, proto.GetString(mixer.Name))

	if error == nil {
		//mixer.RootPackage = pkg
		// Now that we're done resolving, slice off the members! (Ouch)
		mixer.Package = mixer.RootPackage.Package

	} else if error.Code != NOT_FOUND {
		//TODO : Put this into a debug log
		panic(error.Message)
	}

	versionFile := filepath.Join(mixer.DefinitionPath, "..", "..", "JENKINS")
	buildNumber, err := ioutil.ReadFile(versionFile)

	if err == nil {
		mixer.Version = proto.String(proto.GetString(mixer.Version) + "." + strings.Trim(string(buildNumber), "\n\r "))
	}

	return mixer
}

func (m *Mixer) loadDependentMixers() {

	data, err := ioutil.ReadFile(filepath.Join(m.DefinitionPath, "/dependencies.yml"))

	if err != nil {
		// No dependencies
		return
	}

	dependencies := make(map[string]string)
	yaml.Unmarshal(data, &dependencies)

	for fullName, _ := range dependencies {
		m.loadDependentMixer(m.LibraryPath, fullName)
	}

}

func (m *Mixer) loadDependentMixer(buildPath string, name string) {

	// TODO(SJ) : Do I want to check access for these mixers?
	// -- or do I just want a generic 'build' feature access?
	var newMixer *Mixer

	if strings.HasSuffix(name, ".mxr") {
		newMixer = &Mixer{
			Mixer: tp.OpenMixer(filepath.Join(buildPath, name)),
		}
	} else {
		newMixer = BuildMixer(buildPath, name, m.DataPath)
	}

	m.Merge(newMixer)
}

func (m *Mixer) Merge(otherMixer *Mixer) {
	// TODO(SJ) : Make sure there are no name collision in the following unions
	if len(otherMixer.Rewriters) > 0 {

		if len(m.Rewriters) > 0 {
			thisName := proto.GetString(m.Name)
			otherName := proto.GetString(otherMixer.Name)
			panic(fmt.Sprintf("Duplicate sets of rewriters. Mixer (%v) and mixer (%v) both define rewriters.", thisName, otherName))
		}

		m.Rewriters = otherMixer.Rewriters
	}

	// Merge only exists on (tritium) packager.Package
	m.RootPackage.Merge(otherMixer.Package)

	//	m.RootPackage.Dependencies = append(m.RootPackage.Dependencies, proto.GetString(otherMixer.Name) )

}

// In the future, I'll have to search by version too
func findInDataPath(dataPath string, name string) (path *string, err error) {

	prebuiltName := name + ".mxr"

	// Check for pre-built mixer:

	_, err = ioutil.ReadDir(dataPath)

	if err == nil {
		prebuiltPath := filepath.Join(dataPath, prebuiltName)
		_, err = ioutil.ReadFile(prebuiltPath)
		if err == nil {
			return &prebuiltPath, nil
		}
	}

	return nil, errors.New("Couldn't find mixer : " + name)
}

func findInBuildPath(buildPath string, name string) (path *string, err error) {

	// Check for raw mixer directory:
	for _, subPath := range SubPaths {
		thisPath := filepath.Join(buildPath, subPath, name)
		_, err := ioutil.ReadDir(thisPath)

		if err == nil {
			return &thisPath, nil
		}
	}

	return nil, errors.New("Couldn't find mixer : " + name)
}