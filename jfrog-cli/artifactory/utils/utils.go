package utils

import (
	"bufio"
	"errors"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/auth"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/auth/cert"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services"
	clientutils "github.com/AlexeiVainshtein/jfrog-client-go/artifactory/services/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/httpclient"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const repoDetailsUrl = "api/repositories/"

func GetJfrogSecurityDir() (string, error) {
	homeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "security"), nil
}

func GetEncryptedPasswordFromArtifactory(artifactoryAuth auth.ArtifactoryDetails) (string, error) {
	u, err := url.Parse(artifactoryAuth.GetUrl())
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "api/security/encryptedPassword")
	httpClientsDetails := artifactoryAuth.CreateHttpClientDetails()
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return "", err
	}
	transport, err := cert.GetTransportWithLoadedCert(securityDir)
	client := httpclient.NewHttpClient(&http.Client{Transport: transport})
	resp, body, _, err := client.SendGet(u.String(), true, httpClientsDetails)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		return string(body), nil
	}

	if resp.StatusCode == http.StatusConflict {
		message := "\nYour Artifactory server is not configured to encrypt passwords.\n" +
			"You may use \"art config --enc-password=false\""
		return "", errorutils.CheckError(errors.New(message))
	}

	return "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status))
}

func CreateServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := GetJfrogSecurityDir()
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
		SetDryRun(isDryRun).
		SetLogger(log.Logger).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.New(serviceConfig)
}

func ConvertResultItemArrayToDeleteItemArray(resultItems []clientutils.ResultItem) []services.DeleteItem {
	deleteItems := make([]services.DeleteItem, len(resultItems))
	for i, item := range resultItems {
		deleteItems[i] = item
	}
	return deleteItems
}

func isRepoExists(repository string, artDetails auth.ArtifactoryDetails) (bool, error) {
	artHttpDetails := artDetails.CreateHttpClientDetails()
	client := httpclient.NewDefaultHttpClient()
	resp, _, _, err := client.SendGet(artDetails.GetUrl()+repoDetailsUrl+repository, true, artHttpDetails)
	if err != nil {
		return false, errorutils.CheckError(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return true, nil
	}
	return false, nil
}

func CheckIfRepoExists(repository string, artDetails auth.ArtifactoryDetails) error {
	repoExists, err := isRepoExists(repository, artDetails)
	if err != nil {
		return err
	}

	if !repoExists {
		return errorutils.CheckError(errors.New("The repository '" + repository + "' dose not exists."))
	}
	return nil
}

func RunCmdOutput(config CmdConfig) ([]byte, error) {
	for k, v := range config.GetEnv() {
		os.Setenv(k, v)
	}
	cmd := config.GetCmd()
	if config.GetErrWriter() == nil {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = config.GetErrWriter()
		defer config.GetErrWriter().Close()
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return output, nil
}

func RunCmd(config CmdConfig) error {
	for k, v := range config.GetEnv() {
		os.Setenv(k, v)
	}

	cmd := config.GetCmd()
	if config.GetStdWriter() == nil {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = config.GetStdWriter()
		defer config.GetStdWriter().Close()
	}

	if config.GetErrWriter() == nil {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = config.GetErrWriter()
		defer config.GetErrWriter().Close()
	}
	err := cmd.Start()
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = cmd.Wait()
	if err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}

// Executes the command and captures the output.
// Analyze each line to match the provided regex.
func RunCmdWithOutputParser(config CmdConfig, regExpStruct ...*RegExpStruct) error {
	for k, v := range config.GetEnv() {
		os.Setenv(k, v)
	}

	cmd := config.GetCmd()
	cmdReader, err := cmd.StderrPipe()
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer cmdReader.Close()
	scanner := bufio.NewScanner(cmdReader)

	err = cmd.Start()
	if err != nil {
		return errorutils.CheckError(err)
	}

	for scanner.Scan() {
		line := scanner.Text()
		for _, regExp := range regExpStruct {
			regExp.matcher = regExp.RegExp.FindString(line)
			if regExp.matcher != "" {
				regExp.line = line
				line, err = regExp.ExecFunc()
				if err != nil {
					return err
				}
			}
		}
		log.Output(line)
	}
	if scanner.Err() != nil {
		return errorutils.CheckError(scanner.Err())
	}

	err = cmd.Wait()
	if err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}

type CmdConfig interface {
	GetCmd() *exec.Cmd
	GetEnv() map[string]string
	GetStdWriter() io.WriteCloser
	GetErrWriter() io.WriteCloser
}

func GetRegExp(regex string) (*regexp.Regexp, error) {
	regExp, err := regexp.Compile(regex)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	return regExp, nil
}

// Mask the credentials information from the line. The credentials are build as user:password
// For example: http://user:password@127.0.0.1:8081/artifactory/path/to/repo
func (reg *RegExpStruct) MaskCredentials() (string, error) {
	splittedResult := strings.Split(reg.matcher, reg.Separator)
	return strings.Replace(reg.line, reg.matcher, splittedResult[0]+reg.Replacer, 1), nil
}

func (reg *RegExpStruct) ErrorOnNotFound() (string, error) {
	log.Output(reg.line)
	return "", errors.New("404 Not Found")
}

type RegExpStruct struct {
	RegExp    *regexp.Regexp
	matcher   string
	Separator string
	Replacer  string
	line      string
	ExecFunc  func() (string, error)
}
