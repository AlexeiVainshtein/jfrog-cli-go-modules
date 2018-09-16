package bintray

import (
	"github.com/AlexeiVainshtein/jfrog-client-go/bintray/auth"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
)

type Config interface {
	GetUrl() string
	GetKey() string
	GetThreads() int
	GetMinSplitSize() int64
	GetSplitCount() int
	GetMinChecksumDeploy() int64
	IsDryRun() bool
	GetBintrayDetails() auth.BintrayDetails
	GetLogger() log.Log
}
