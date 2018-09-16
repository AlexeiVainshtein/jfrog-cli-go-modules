package services

import (
	"encoding/json"
	"errors"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/missioncontrol/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/cliutils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"net/http"
	"github.com/AlexeiVainshtein/jfrog-client-go/httpclient"
)

func DetachLic(service_name string, flags *DetachLicFlags) error {
	bucketId := flags.BucketId
	postContent := utils.LicenseRequestContent{
		Name:   service_name}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		return errorutils.CheckError(errors.New("Failed to marshal json. " + cliutils.GetDocumentationMessage()))
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v3/detach_lic/buckets/" + bucketId
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	client := httpclient.NewDefaultHttpClient()
	resp, body, err := client.SendDelete(missionControlUrl, requestContent, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	log.Debug("Mission Control response: " + resp.Status)
	return nil
}

type DetachLicFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive           bool
	BucketId              string
}
