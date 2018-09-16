package utils

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/bintray/commands"
	"github.com/AlexeiVainshtein/jfrog-client-go/bintray"
	"github.com/AlexeiVainshtein/jfrog-client-go/bintray/auth"
	"github.com/AlexeiVainshtein/jfrog-client-go/bintray/services"
	"github.com/AlexeiVainshtein/jfrog-client-go/bintray/services/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/io/fileutils"
	"path/filepath"
)

// Download file from Bintray.
// downloadPath: Bintray download path in the following format: subject/repo/path/version/filename.
// filename: the file full name.
// targetPath: local download target path.
func DownloadFromBintrayIfNeeded(downloadPath, filename, targetPath string) error {
	targetFile := filepath.Join(targetPath, filename)
	exists, err := fileutils.IsFileExists(targetFile)
	if exists || err != nil {
		return err
	}

	bintrayConfig := auth.NewBintrayDetails()
	config := bintray.NewConfigBuilder().SetBintrayDetails(bintrayConfig).Build()

	pathDetails, err := utils.CreatePathDetails(downloadPath)
	if err != nil {
		return err
	}

	params := &services.DownloadFileParams{}
	params.PathDetails = pathDetails
	params.TargetPath = targetFile
	params.Flat = true

	_, _, err = commands.DownloadFile(config, params)
	return err
}
