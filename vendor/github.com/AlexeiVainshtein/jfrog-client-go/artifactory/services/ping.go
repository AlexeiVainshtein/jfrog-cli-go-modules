package services

import (
	"errors"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/auth"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/httpclient"
	clientutils "github.com/AlexeiVainshtein/jfrog-client-go/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"net/http"
)

type PingService struct {
	httpClient *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewPingService(client *httpclient.HttpClient) *PingService {
	return &PingService{httpClient: client}
}

func (ps *PingService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return ps.ArtDetails
}

func (ps *PingService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	ps.ArtDetails = rt
}

func (ps *PingService) GetJfrogHttpClient() *httpclient.HttpClient {
	return ps.httpClient
}

func (ps *PingService) IsDryRun() bool {
	return false
}

func (ps *PingService) Ping() ([]byte, error) {
	url, err := utils.BuildArtifactoryUrl(ps.ArtDetails.GetUrl(), "api/system/ping", nil)
	if err != nil {
		return nil, err
	}
	resp, respBody, _, err := ps.httpClient.SendGet(url, true, ps.ArtDetails.CreateHttpClientDetails())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(respBody)))
	}
	log.Debug("Artifactory response: ", resp.Status)
	return respBody, nil
}