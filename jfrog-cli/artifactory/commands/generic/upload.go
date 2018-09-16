package generic

import (
	"errors"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/spec"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/buildinfo"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
	clientutils "github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/io/fileutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"os"
	"strconv"
	"strings"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(uploadSpec *spec.SpecFiles, flags *UploadConfiguration) (successCount, failCount int, err error) {
	certPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		return 0, 0, err
	}
	minChecksumDeploySize, err := getMinChecksumDeploySize()
	if err != nil {
		return 0, 0, err
	}
	servicesConfig, err := createUploadServiceConfig(flags.ArtDetails, flags, certPath, minChecksumDeploySize)
	if err != nil {
		return 0, 0, err
	}
	servicesManager, err := artifactory.New(servicesConfig)
	if err != nil {
		return 0, 0, err
	}
	isCollectBuildInfo := len(flags.BuildName) > 0 && len(flags.BuildNumber) > 0
	if isCollectBuildInfo && !flags.DryRun {
		if err := utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); err != nil {
			return 0, 0, err
		}
		for i := 0; i < len(uploadSpec.Files); i++ {
			addBuildProps(&uploadSpec.Get(i).Props, flags.BuildName, flags.BuildNumber)
		}
	}

	uploadParamImp := createBaseUploadParams(flags)
	var filesInfo []clientutils.FileInfo
	var errorOccurred = false
	for i := 0; i < len(uploadSpec.Files); i++ {
		params, err := uploadSpec.Get(i).ToArtifatoryUploadParams()
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		uploadParamImp.ArtifactoryCommonParams = params
		flat, err := uploadSpec.Get(i).IsFlat(true)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		uploadParamImp.Flat = flat
		explode, err := uploadSpec.Get(i).IsExplode(false)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		uploadParamImp.ExplodeArchive = explode
		artifacts, uploaded, failed, err := servicesManager.UploadFiles(uploadParamImp)
		filesInfo = append(filesInfo, artifacts...)
		failCount += failed
		successCount += uploaded
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
	}
	if errorOccurred {
		err = errors.New("Upload finished with errors. Please review the logs")
		return
	}
	if failCount > 0 {
		return
	}
	if isCollectBuildInfo && !flags.DryRun {
		buildArtifacts := convertFileInfoToBuildArtifacts(filesInfo)
		populateFunc := func(partial *buildinfo.Partial) {
			partial.Artifacts = buildArtifacts
		}
		err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return
}

func convertFileInfoToBuildArtifacts(filesInfo []clientutils.FileInfo) []buildinfo.Artifact {
	buildArtifacts := make([]buildinfo.Artifact, len(filesInfo))
	for i, fileInfo := range filesInfo {
		artifact := buildinfo.Artifact{Checksum: &buildinfo.Checksum{}}
		artifact.Sha1 = fileInfo.Sha1
		artifact.Md5 = fileInfo.Md5
		filename, _ := fileutils.GetFileAndDirFromPath(fileInfo.LocalPath)
		artifact.Name = filename
		buildArtifacts[i] = artifact
	}
	return buildArtifacts
}

func createUploadServiceConfig(artDetails *config.ArtifactoryDetails, flags *UploadConfiguration, certPath string, minChecksumDeploySize int64) (artifactory.Config, error) {
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	servicesConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetDryRun(flags.DryRun).
		SetCertificatesPath(certPath).
		SetMinChecksumDeploy(minChecksumDeploySize).
		SetThreads(flags.Threads).
		SetLogger(log.Logger).
		Build()
	return servicesConfig, err
}

func createBaseUploadParams(flags *UploadConfiguration) *services.UploadParamsImp {
	uploadParamImp := &services.UploadParamsImp{}
	uploadParamImp.Deb = flags.Deb
	uploadParamImp.Symlink = flags.Symlink
	uploadParamImp.Retries = flags.Retries
	return uploadParamImp
}

func getMinChecksumDeploySize() (int64, error) {
	minChecksumDeploySize := os.Getenv("JFROG_CLI_MIN_CHECKSUM_DEPLOY_SIZE_KB")
	if minChecksumDeploySize == "" {
		return 10240, nil
	}
	minSize, err := strconv.ParseInt(minChecksumDeploySize, 10, 64)
	err = errorutils.CheckError(err)
	if err != nil {
		return 0, err
	}
	return minSize * 1000, nil
}

func addBuildProps(props *string, buildName, buildNumber string) error {
	if buildName == "" || buildNumber == "" {
		return nil
	}
	buildProps, err := utils.CreateBuildProperties(buildName, buildNumber)
	if err != nil {
		return err
	}

	if len(*props) > 0 && !strings.HasSuffix(*props, ";") && len(buildProps) > 0 {
		*props += ";"
	}
	*props += buildProps
	return nil
}

type UploadConfiguration struct {
	Deb                   string
	Threads               int
	MinChecksumDeploySize int64
	BuildName             string
	BuildNumber           string
	DryRun                bool
	Symlink               bool
	ExplodeArchive        bool
	ArtDetails            *config.ArtifactoryDetails
	Retries               int
}
