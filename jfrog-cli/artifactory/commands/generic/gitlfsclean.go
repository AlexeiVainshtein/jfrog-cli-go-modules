package generic

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
	clientutils "github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
)

func PrepareGitLfsClean(flags *GitLfsCleanConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return nil, err
	}
	return servicesManager.GetUnreferencedGitLfsFiles(flags)
}

func DeleteLfsFilesFromArtifactory(files []clientutils.ResultItem, flags *GitLfsCleanConfiguration) error {
	log.Info("Deleting", len(files), "files from", flags.Repo, "...")
	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	deleteItems := utils.ConvertResultItemArrayToDeleteItemArray(files)
	_, err = servicesManager.DeleteFiles(deleteItems)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

type GitLfsCleanConfiguration struct {
	*services.GitLfsCleanParamsImpl
	ArtDetails *config.ArtifactoryDetails
	Quiet      bool
	DryRun     bool
}
