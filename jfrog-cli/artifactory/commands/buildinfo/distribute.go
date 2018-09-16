package buildinfo

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
)

func Distribute(flags *BuildDistributionConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.DistributeBuild(flags.BuildDistributionParamsImpl)
}

type BuildDistributionConfiguration struct {
	*services.BuildDistributionParamsImpl
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}
