package buildinfo

import (
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
)

func Clean(buildName, buildNumber string) error {
	log.Info("Cleaning build info...")
	err := utils.RemoveBuildDir(buildName, buildNumber)
	if err != nil {
		return err
	}
	log.Info("Cleaned build info", buildName+"/"+buildNumber+".")
	return nil
}
