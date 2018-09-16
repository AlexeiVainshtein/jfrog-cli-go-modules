package config

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/buger/jsonparser"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/cliutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/artifactory/auth"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/io/fileutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/prompt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"fmt"
)

// This is the default server id. It is used when adding a server config without providing a server ID
const (
	DefaultServerId   = "Default-Server"
	JfrogHomeEnv      = "JFROG_CLI_HOME"
	JfrogConfigFile   = "jfrog-cli.conf"
	JfrogDependencies = "dependencies"
)

func IsArtifactoryConfExists() (bool, error) {
	conf, err := readConf()
	if err != nil {
		return false, err
	}
	return conf.Artifactory != nil && len(conf.Artifactory) > 0, nil
}

func IsMissionControlConfExists() (bool, error) {
	conf, err := readConf()
	if err != nil {
		return false, err
	}
	return conf.MissionControl != nil, nil
}

func IsBintrayConfExists() (bool, error) {
	conf, err := readConf()
	if err != nil {
		return false, err
	}
	return conf.Bintray != nil, nil
}

func GetArtifactorySpecificConfig(serverId string) (*ArtifactoryDetails, error) {
	conf, err := readConf()
	if err != nil {
		return nil, err
	}
	details := conf.Artifactory
	if details == nil || len(details) == 0 {
		return new(ArtifactoryDetails), nil
	}
	if len(serverId) == 0 {
		return GetDefaultArtifactoryConf(details)
	}
	return GetArtifactoryConfByServerId(serverId, details)
}

func GetDefaultArtifactoryConf(configs []*ArtifactoryDetails) (*ArtifactoryDetails, error) {
	if len(configs) == 0 {
		details := new(ArtifactoryDetails)
		details.IsDefault = true
		return details, nil
	}
	for _, conf := range configs {
		if conf.IsDefault == true {
			return conf, nil
		}
	}
	return nil, errorutils.CheckError(errors.New("Couldn't find default server."))
}

// Returns the configured server or error if the server id not found
func GetArtifactoryConfByServerId(serverName string, configs []*ArtifactoryDetails) (*ArtifactoryDetails, error) {
	for _, conf := range configs {
		if conf.ServerId == serverName {
			return conf, nil
		}
	}
	return nil, errorutils.CheckError(errors.New(fmt.Sprintf("Server id '%s' dose not exists.", serverName)))
}

func GetAndRemoveConfiguration(serverName string, configs []*ArtifactoryDetails) (*ArtifactoryDetails, []*ArtifactoryDetails) {
	for i, conf := range configs {
		if conf.ServerId == serverName {
			configs = append(configs[:i], configs[i+1:]...)
			return conf, configs
		}
	}
	return nil, configs
}

func GetAllArtifactoryConfigs() ([]*ArtifactoryDetails, error) {
	conf, err := readConf()
	if err != nil {
		return nil, err
	}
	details := conf.Artifactory
	if details == nil {
		return make([]*ArtifactoryDetails, 0), nil
	}
	return details, nil
}

func ReadMissionControlConf() (*MissionControlDetails, error) {
	conf, err := readConf()
	if err != nil {
		return nil, err
	}
	details := conf.MissionControl
	if details == nil {
		return new(MissionControlDetails), nil
	}
	return details, nil
}

func ReadBintrayConf() (*BintrayDetails, error) {
	conf, err := readConf()
	if err != nil {
		return nil, err
	}
	details := conf.Bintray
	if details == nil {
		return new(BintrayDetails), nil
	}
	return details, nil
}

func SaveArtifactoryConf(details []*ArtifactoryDetails) error {
	conf, err := readConf()
	if err != nil {
		return err
	}
	conf.Artifactory = details
	return saveConfig(conf)
}

func SaveMissionControlConf(details *MissionControlDetails) error {
	conf, err := readConf()
	if err != nil {
		return err
	}
	conf.MissionControl = details
	return saveConfig(conf)
}

func SaveBintrayConf(details *BintrayDetails) error {
	config, err := readConf()
	if err != nil {
		return err
	}
	config.Bintray = details
	return saveConfig(config)
}

func saveConfig(config *ConfigV1) error {
	config.Version = cliutils.GetConfigVersion()
	b, err := json.Marshal(&config)
	if err != nil {
		return errorutils.CheckError(err)
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	if err != nil {
		return errorutils.CheckError(err)
	}
	path, err := getConfFilePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(content.String()), 0600)
	if err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}

func readConf() (*ConfigV1, error) {
	confFilePath, err := getConfFilePath()
	if err != nil {
		return nil, err
	}
	config := new(ConfigV1)
	exists, err := fileutils.IsFileExists(confFilePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return config, nil
	}
	content, err := fileutils.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return new(ConfigV1), nil
	}
	content, err = convertIfNecessary(content)
	err = json.Unmarshal(content, &config)
	return config, err
}

// The configuration schema can change between versions, therefore we need to convert old versions to the new schema.
func convertIfNecessary(content []byte) ([]byte, error) {
	version, err := jsonparser.GetString(content, "Version")
	if err != nil {
		if err.Error() == "Key path not found" {
			version = "0"
		} else {
			return nil, errorutils.CheckError(err)
		}
	}
	switch version {
	case "0":
		result := new(ConfigV1)
		configV0 := new(ConfigV0)
		err = json.Unmarshal(content, &configV0)
		if errorutils.CheckError(err) != nil {
			return nil, err
		}
		result = configV0.Convert()
		err = saveConfig(result)
		content, err = json.Marshal(&result)
	}
	return content, err
}

func GetJfrogHomeDir() (string, error) {
	if os.Getenv(JfrogHomeEnv) != "" {
		return path.Join(os.Getenv(JfrogHomeEnv), ".jfrog"), nil
	}

	userDir := fileutils.GetHomeDir()
	if userDir == "" {
		err := errorutils.CheckError(errors.New("Couldn't find home directory. Make sure your HOME environment variable is set."))
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(userDir, ".jfrog"), nil
}

func GetJfrogDependenciesPath() (string, error) {
	jfrogHome, err := GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(jfrogHome, JfrogDependencies), nil
}

func getConfFilePath() (string, error) {
	confPath, err := GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	os.MkdirAll(confPath, 0777)
	return filepath.Join(confPath, JfrogConfigFile), nil
}

type ConfigV1 struct {
	Artifactory    []*ArtifactoryDetails  `json:"artifactory"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"MissionControl,omitempty"`
	Version        string                 `json:"Version,omitempty"`
}

type ConfigV0 struct {
	Artifactory    *ArtifactoryDetails    `json:"artifactory,omitempty"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"MissionControl,omitempty"`
}

func (o *ConfigV0) Convert() *ConfigV1 {
	config := new(ConfigV1)
	config.Bintray = o.Bintray
	config.MissionControl = o.MissionControl
	if o.Artifactory != nil {
		o.Artifactory.IsDefault = true
		o.Artifactory.ServerId = DefaultServerId
		config.Artifactory = []*ArtifactoryDetails{o.Artifactory}
	}
	return config
}

type ArtifactoryDetails struct {
	Url            string            `json:"url,omitempty"`
	User           string            `json:"user,omitempty"`
	Password       string            `json:"password,omitempty"`
	SshKeyPath     string            `json:"sshKeyPath,omitempty"`
	SshPassphrase  string            `json:"SshPassphrase,omitempty"`
	SshAuthHeaders map[string]string `json:"SshAuthHeaders,omitempty"`
	ServerId       string            `json:"serverId,omitempty"`
	IsDefault      bool              `json:"isDefault,omitempty"`
	// Deprecated, use password option instead.
	ApiKey         string            `json:"apiKey,omitempty"`
}

type BintrayDetails struct {
	ApiUrl            string `json:"-"`
	DownloadServerUrl string `json:"-"`
	User              string `json:"user,omitempty"`
	Key               string `json:"key,omitempty"`
	DefPackageLicense string `json:"defPackageLicense,omitempty"`
}

type MissionControlDetails struct {
	Url      string `json:"url,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

func (artifactoryDetails *ArtifactoryDetails) IsEmpty() bool {
	return len(artifactoryDetails.Url) == 0
}

func (artifactoryDetails *ArtifactoryDetails) SetApiKey(apiKey string) {
	artifactoryDetails.ApiKey = apiKey
}

func (artifactoryDetails *ArtifactoryDetails) SetUser(username string) {
	artifactoryDetails.User = username
}

func (artifactoryDetails *ArtifactoryDetails) SetPassword(password string) {
	artifactoryDetails.Password = password
}

func (artifactoryDetails *ArtifactoryDetails) GetApiKey() string {
	return artifactoryDetails.ApiKey
}

func (artifactoryDetails *ArtifactoryDetails) GetUrl() string {
	return artifactoryDetails.Url
}

func (artifactoryDetails *ArtifactoryDetails) GetUser() string {
	return artifactoryDetails.User
}

func (artifactoryDetails *ArtifactoryDetails) GetPassword() string {
	return artifactoryDetails.Password
}

func (artifactoryDetails *ArtifactoryDetails) SshAuthHeaderSet() bool {
	return len(artifactoryDetails.SshAuthHeaders) > 0
}

func (artifactoryDetails *ArtifactoryDetails) sshAuthenticationRequired() bool {
	return !artifactoryDetails.SshAuthHeaderSet() && fileutils.IsSshUrl(artifactoryDetails.Url)
}

func (artifactoryDetails *ArtifactoryDetails) CreateArtAuthConfig() (auth.ArtifactoryDetails, error) {
	artAuth := auth.NewArtifactoryDetails()
	artAuth.SetUrl(artifactoryDetails.Url)
	artAuth.SetSshAuthHeaders(artifactoryDetails.SshAuthHeaders)
	artAuth.SetApiKey(artifactoryDetails.ApiKey)
	artAuth.SetUser(artifactoryDetails.User)
	artAuth.SetPassword(artifactoryDetails.Password)
	if artifactoryDetails.sshAuthenticationRequired() {
		var sshKey, sshPassphrase []byte
		var err error
		if len(artifactoryDetails.SshKeyPath) > 0 {
			sshKey, sshPassphrase, err = readSshKeyAndPassphrase(artifactoryDetails.SshKeyPath, artifactoryDetails.SshPassphrase)
			if err != nil {
				return nil, err
			}
		}
		err = artAuth.AuthenticateSsh(sshKey, sshPassphrase)
		if err != nil {
			return nil, err
		}
	}
	return artAuth, nil
}

func readSshKeyAndPassphrase(sshKeyPath, sshPassphrase string) ([]byte, []byte, error) {
	sshKey, err := ioutil.ReadFile(utils.ReplaceTildeWithUserHome(sshKeyPath))
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}
	if len(sshPassphrase) == 0 {
		encryptedKey, err := isEncrypted(sshKey)
		if errorutils.CheckError(err) != nil {
			return nil, nil, err
		}
		if encryptedKey {
			sshPassphrase, err = readSshPassphrase(sshKeyPath)
			if errorutils.CheckError(err) != nil {
				return nil, nil, err
			}
		}
	}

	return sshKey, []byte(sshPassphrase), err
}

func readSshPassphrase(sshKeyPath string) (string, error) {
	offerConfig, err := cliutils.GetBoolEnvValue("JFROG_CLI_OFFER_CONFIG", true)
	if err != nil || !offerConfig {
		return "", err
	}
	simplePrompt := &prompt.Simple{
		Msg:   "Enter passphrase for key '" + sshKeyPath + "': ",
		Mask:  true,
		Label: "sshPassphrase",
	}
	if err = simplePrompt.Read(); err != nil {
		return "", err
	}
	return simplePrompt.GetResults().GetString("sshPassphrase"), nil
}

func isEncrypted(buffer []byte) (bool, error) {
	block, _ := pem.Decode(buffer)
	if block == nil {
		return false, errors.New("SSH: no key found")
	}
	return strings.Contains(block.Headers["Proc-Type"], "ENCRYPTED"), nil
}

func (missionControlDetails *MissionControlDetails) SetUser(username string) {
	missionControlDetails.User = username
}

func (missionControlDetails *MissionControlDetails) SetPassword(password string) {
	missionControlDetails.Password = password
}

func (missionControlDetails *MissionControlDetails) GetUser() string {
	return missionControlDetails.User
}

func (missionControlDetails *MissionControlDetails) GetPassword() string {
	return missionControlDetails.Password
}
