package buildinfo

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
)

func Promote(flags *BuildPromotionConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	return servicesManager.PromoteBuild(flags.PromotionParamsImpl)
}

type BuildPromotionConfiguration struct {
	*services.PromotionParamsImpl
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
}
