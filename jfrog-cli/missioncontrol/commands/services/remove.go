package services

import (
	"errors"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/missioncontrol/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"github.com/AlexeiVainshtein/jfrog-client-go/httpclient"
)

func Remove(serviceName string, flags *RemoveFlags) error {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v3/services/" + serviceName
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	client := httpclient.NewDefaultHttpClient()
	resp, body, err := client.SendDelete(missionControlUrl, nil, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	log.Debug("Mission Control response: " + resp.Status)
	return nil
}

type RemoveFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive           bool
}
