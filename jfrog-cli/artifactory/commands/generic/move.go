package generic

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/spec"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
)

// Moves the artifacts using the specified move pattern.
func Move(moveSpec *spec.SpecFiles, flags *MoveConfiguration) (successCount, failCount int, err error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return
	}
	for i := 0; i < len(moveSpec.Files); i++ {
		params, err := moveSpec.Get(i).ToArtifatoryMoveCopyParams()
		if err != nil {
			log.Error(err)
			continue
		}
		flat, err := moveSpec.Get(i).IsFlat(false)
		if err != nil {
			log.Error(err)
			continue
		}
		partialSuccess, partialFailed, err := servicesManager.Move(&services.MoveCopyParamsImpl{ArtifactoryCommonParams: params, Flat: flat})
		successCount += partialSuccess
		failCount += partialFailed
		if err != nil {
			log.Error(err)
			continue
		}
	}
	return
}

type MoveConfiguration struct {
	DryRun                bool
	ArtDetails            *config.ArtifactoryDetails
}