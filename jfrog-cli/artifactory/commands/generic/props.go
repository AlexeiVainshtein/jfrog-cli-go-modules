package generic

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/spec"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
	clientutils "github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
)

func SetProps(spec *spec.SpecFiles, props string, threads int, artDetails *config.ArtifactoryDetails) (successCount, failCount int, err error) {
	servicesManager, err := createPropsServiceManager(threads, artDetails)
	if err != nil {
		return 0, 0, err
	}

	resultItems := searchItems(spec, servicesManager)
	success, err := servicesManager.SetProps(&services.PropsParamsImpl{Items: resultItems, Props: props})
	return success, len(resultItems) - success, err
}

func DeleteProps(spec *spec.SpecFiles, props string, threads int, artDetails *config.ArtifactoryDetails) (successCount, failCount int, err error) {
	servicesManager, err := createPropsServiceManager(threads, artDetails)
	if err != nil {
		return 0, 0, err
	}

	resultItems := searchItems(spec, servicesManager)
	success, err := servicesManager.DeleteProps(&services.PropsParamsImpl{Items: resultItems, Props: props})
	return success, len(resultItems) - success, err
}

func createPropsServiceManager(threads int, artDetails *config.ArtifactoryDetails) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetCertificatesPath(certPath).
		SetLogger(log.Logger).
		SetThreads(threads).
		Build()

	return artifactory.New(serviceConfig)
}

func searchItems(spec *spec.SpecFiles, servicesManager *artifactory.ArtifactoryServicesManager) (resultItems []clientutils.ResultItem) {
	for i := 0; i < len(spec.Files); i++ {
		params, err := spec.Get(i).ToArtifatorySetPropsParams()
		if err != nil {
			log.Error(err)
			continue
		}
		currentResultItems, err := servicesManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: params})
		if err != nil {
			log.Error(err)
			continue
		}
		resultItems = append(resultItems, currentResultItems...)
	}
	return
}
