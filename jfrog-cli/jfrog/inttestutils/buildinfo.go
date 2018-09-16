package inttestutils

import (
	"encoding/json"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/buildinfo"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/io/httputils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"net/http"
	"testing"
	"github.com/AlexeiVainshtein/jfrog-client-go/httpclient"
)

func GetBuildInfo(artifactoryUrl, buildName, buildNumber string, t *testing.T, artHttpDetails httputils.HttpClientDetails) buildinfo.BuildInfo {
	client := httpclient.NewDefaultHttpClient()
	_, body, _, err := client.SendGet(artifactoryUrl+"api/build/"+buildName+"/"+buildNumber, true, artHttpDetails)
	if err != nil {
		t.Error(err)
	}

	var buildInfoJson struct {
		BuildInfo buildinfo.BuildInfo `json:"buildInfo,omitempty"`
	}
	if err := json.Unmarshal(body, &buildInfoJson); err != nil {
		t.Error(err)
	}
	return buildInfoJson.BuildInfo
}

func DeleteBuild(artifactoryUrl, buildName string, artHttpDetails httputils.HttpClientDetails) {
	client := httpclient.NewDefaultHttpClient()
	resp, body, err := client.SendDelete(artifactoryUrl+"api/build/"+buildName+"?deleteAll=1", nil, artHttpDetails)
	if err != nil {
		log.Error(err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Error(resp.Status)
		log.Error(string(body))
	}
}
