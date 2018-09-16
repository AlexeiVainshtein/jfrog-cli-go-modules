package buildinfo

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
)

func BuildDiscard(flags *BuildDiscardConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DiscardBuilds(flags.DiscardBuildsParamsImpl)
}

type BuildDiscardConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	*services.DiscardBuildsParamsImpl
}
