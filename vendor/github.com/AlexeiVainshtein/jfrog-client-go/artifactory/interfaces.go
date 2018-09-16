package artifactory

import (
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/auth"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
)

type Config interface {
	GetUrl() string
	GetPassword() string
	GetApiKey() string
	GetCertifactesPath() string
	GetThreads() int
	GetMinSplitSize() int64
	GetSplitCount() int
	GetMinChecksumDeploy() int64
	IsDryRun() bool
	GetArtDetails() auth.ArtifactoryDetails
	GetLogger() log.Log
}

type ArtifactoryServicesSetter interface {
	SetThread(threads int)
	SetArtDetails(artDetails auth.ArtifactoryDetails)
	SetDryRun(isDryRun bool)
}
