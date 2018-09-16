package golang

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	goutils "github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils/golang"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils/golang/project"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"os"
	"os/exec"
	"strings"
)

func Publish(publishPackage bool, dependencies, targetRepo, version, buildName, buildNumber string, details *config.ArtifactoryDetails) (succeeded, failed int, err error) {
	err = validatePrerequisites()
	if err != nil {
		return
	}

	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err = utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return
		}
	}

	goProject, err := project.Load(version)
	if err != nil {
		return
	}

	// Publish the package to Artifactory
	if publishPackage {
		err = goProject.PublishPackage(targetRepo, buildName, buildNumber, details)
		if err != nil {
			return
		}
	}

	// Publish the package dependencies to Artifactory
	depsList := strings.Split(dependencies, ",")
	if len(depsList) > 0 {
		succeeded, failed, err = goProject.PublishDependencies(targetRepo, details, depsList)
	}
	if err != nil {
		return
	}
	if publishPackage {
		succeeded++
	}

	// Publish the build-info to Artifactory
	if isCollectBuildInfo {
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(true))
	}

	return
}

func ExecuteGo(noRegistry bool, goArg, targetRepo, buildName, buildNumber string, details *config.ArtifactoryDetails) error {
	isCollectBuildInfo := len(buildName) > 0 && len(buildNumber) > 0
	if isCollectBuildInfo {
		err := utils.SaveBuildGeneralDetails(buildName, buildNumber)
		if err != nil {
			return err
		}
	}

	if !noRegistry {
		goutils.SetGoProxyEnvVar(details, targetRepo)
	}
	err := goutils.RunGo(goArg)
	if err != nil {
		if !noRegistry && strings.EqualFold(err.Error(), "404 Not Found") {
			// Need to run Go without Artifactory to resolve all dependencies.
			log.Info("Received", err.Error(), "from Artifactory. Trying download the dependencies from the VCS and upload them to Artifactory...")
			err = downloadAndPublish(targetRepo, details)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if isCollectBuildInfo {
		// The version is not necessary because we are collecting the dependencies only.
		goProject, err := project.Load("-")
		if err != nil {
			return err
		}
		err = utils.SaveBuildInfo(buildName, buildNumber, goProject.BuildInfo(false))
	}

	return err
}

func validatePrerequisites() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

// Download the dependencies from VCS and publish them to Artifactory.
func downloadAndPublish(targetRepo string, details *config.ArtifactoryDetails) error {
	err := os.Unsetenv(goutils.GOPROXY)
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = goutils.DownloadDependenciesDirectly()
	if err != nil {
		return err
	}
	// Publish the dependencies.
	_, _, err = Publish(false, "ALL", targetRepo, "", "", "", details)
	if err != nil {
		return err
	}

	return nil
}
