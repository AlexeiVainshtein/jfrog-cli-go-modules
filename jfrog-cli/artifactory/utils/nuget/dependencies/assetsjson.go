package dependencies

import (
	"encoding/json"
	"fmt"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/buildinfo"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/io/fileutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"io/ioutil"
	"path/filepath"
	"strings"
	"errors"
)

var assetsFilePath = filepath.Join("obj", "project.assets.json")

// Register project.assets.json extractor
func init() {
	register(&assetsExtractor{})
}

// project.assets.json dependency extractor
type assetsExtractor struct {
	assets *assets
}

func (extractor *assetsExtractor) IsCompatible(projectName, projectRoot string) (bool, error) {
	assetsFilePath := filepath.Join(projectRoot, assetsFilePath)
	exists, err := fileutils.IsFileExists(assetsFilePath)
	if exists {
		log.Debug("Found", assetsFilePath, "file for project:", projectName)
		return true, err
	}
	return false, err
}

func (extractor *assetsExtractor) DirectDependencies() ([]string, error) {
	return extractor.assets.getDirectDependencies(), nil
}

func (extractor *assetsExtractor) AllDependencies() (map[string]*buildinfo.Dependency, error) {
	return extractor.assets.getAllDependencies()
}

func (extractor *assetsExtractor) ChildrenMap() (map[string][]string, error) {
	return extractor.assets.getChildrenMap(), nil
}

// Create new assets json extractor.
func (extractor *assetsExtractor) new(projectName, projectRoot string) (Extractor, error) {
	newExtractor := &assetsExtractor{}
	assetsFilePath := filepath.Join(projectRoot, assetsFilePath)
	content, err := ioutil.ReadFile(assetsFilePath)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	assets := &assets{}
	err = json.Unmarshal(content, assets)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	newExtractor.assets = assets
	return newExtractor, nil
}

func (assets *assets) getChildrenMap() map[string][]string {
	dependenciesRelations := map[string][]string{}
	for _, dependencies := range assets.Targets {
		for dependencyId, targetDependencies := range dependencies {
			var transitive []string
			for transitiveName := range targetDependencies.Dependencies {
				transitive = append(transitive, strings.ToLower(transitiveName))
			}
			dependencyName := getDependencyName(dependencyId)
			dependenciesRelations[dependencyName] = transitive
		}
	}
	return dependenciesRelations
}

func (assets *assets) getDirectDependencies() []string {
	var directDependencies []string
	for _, framework := range assets.Project.Frameworks {
		for dependencyName := range framework.Dependencies {
			directDependencies = append(directDependencies, strings.ToLower(dependencyName))
		}
	}
	return directDependencies
}

func (assets *assets) getAllDependencies() (map[string]*buildinfo.Dependency, error) {
	dependencies := map[string]*buildinfo.Dependency{}
	packagesPath := assets.Project.Restore.PackagesPath
	for dependencyId, library := range assets.Libraries {
		nupkgFileName, err := library.getNupkgFileName()
		if err != nil {
			return nil, err
		}
		nupkgFilePath := filepath.Join(packagesPath, library.Path, nupkgFileName)
		exists, err := fileutils.IsFileExists(nupkgFilePath)
		if err != nil {
			return nil, err
		}
		if !exists {
			if assets.isPackagePartOfTargetDependencies(library.Path) {
				log.Warn("The file", nupkgFilePath, "doesn't exist in the NuGet cache directory but it does exist as a target in the assets files. Skipping adding this file to the build info.")
				continue
			}
			return nil, errorutils.CheckError(errors.New("The file " + nupkgFilePath + " doesn't exist in the NuGet cache directory."))
		}
		fileDetails, err := fileutils.GetFileDetails(nupkgFilePath)
		if err != nil {
			return nil, err
		}

		dependencyName := getDependencyName(dependencyId)
		dependencies[dependencyName] = &buildinfo.Dependency{Id: dependencyId, Checksum: &buildinfo.Checksum{Sha1: fileDetails.Checksum.Sha1, Md5: fileDetails.Checksum.Md5}}
	}

	return dependencies, nil
}

// If the package is included in the targets section of the assets.json file,
// then this is a .NET dependency that shouldn't be included in the build-info dependencies list
// (it come with the SDK).
// Those files are located in the following path: C:\Program Files\dotnet\sdk\NuGetFallbackFolder
func (assets *assets) isPackagePartOfTargetDependencies(nugetPackageName string) bool {
	for _, dependencies := range assets.Targets {
		for dependencyId := range dependencies {
			// The package names in the targets section of the assets.json file are
			// case insensitive.
			if strings.EqualFold(dependencyId, nugetPackageName) {
				return true
			}
		}
	}
	return false
}

func getDependencyName(dependencyId string) string {
	return strings.ToLower(dependencyId)[0:strings.Index(dependencyId, "/")]
}

// Assets json objects for unmarshalling
type assets struct {
	Version   int
	Targets   map[string]map[string]targetDependency `json:"targets,omitempty"`
	Libraries map[string]library                     `json:"libraries,omitempty"`
	Project   project                                `json:"project"`
}

type targetDependency struct {
	Dependencies map[string]string `json:"dependencies,omitempty"` // Transitive dependencies
}

type library struct {
	Path  string   `json:"path,omitempty"`
	Files []string `json:"files,omitempty"`
}

func (library *library) getNupkgFileName() (string, error) {
	for _, fileName := range library.Files {
		if strings.HasSuffix(fileName, "nupkg.sha512") {
			return strings.TrimSuffix(fileName, ".sha512"), nil
		}
	}
	return "", errorutils.CheckError(fmt.Errorf("Could not find nupkg file name for: %s", library.Path))
}

type project struct {
	Version    string               `json:"version,omitempty"`
	Restore    restore              `json:"restore"`
	Frameworks map[string]framework `json:"frameworks,omitempty"`
}

type restore struct {
	PackagesPath string `json:"packagesPath"`
}

type framework struct {
	Dependencies map[string]dependency `json:"dependencies,omitempty"` // Direct dependencies
}

type dependency struct {
	Target  string `json:"target"`
	Version string `json:"version,omitempty"`
}
