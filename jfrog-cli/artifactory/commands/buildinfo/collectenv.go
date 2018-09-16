package buildinfo

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/buildinfo"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"os"
	"strings"
)

func CollectEnv(buildName, buildNumber string) error {
	log.Info("Collecting environment variables...")
	err := utils.SaveBuildGeneralDetails(buildName, buildNumber)
	if err != nil {
		return err
	}
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Env = getEnvVariables()
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	if err != nil {
		return err
	}
	log.Info("Collected environment variables for", buildName+"/"+buildNumber+".")
	return nil
}

func getEnvVariables() buildinfo.Env {
	m := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if len(pair[0]) != 0 {
			m["buildInfo.env."+pair[0]] = pair[1]
		}
	}
	return m
}
