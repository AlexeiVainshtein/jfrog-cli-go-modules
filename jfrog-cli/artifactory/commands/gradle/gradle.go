package gradle

import (
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/artifactory/utils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/io/fileutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"fmt"
)

const gradleExtractorDependencyVersion = "4.7.5"
const gradleInitScriptTemplate = "gradle.init"

const usePlugin = "useplugin"
const useWrapper = "usewrapper"
const gradleBuildInfoProperties = "BUILDINFO_PROPFILE"

func Gradle(tasks, configPath string, configuration *utils.BuildConfiguration) error {
	gradleDependenciesDir, gradlePluginFilename, err := downloadGradleDependencies()
	if err != nil {
		return err
	}
	gradleRunConfig, err := createGradleRunConfig(tasks, configPath, configuration, gradleDependenciesDir, gradlePluginFilename)
	if err != nil {
		return err
	}
	defer os.Remove(gradleRunConfig.env[gradleBuildInfoProperties])
	if err := utils.RunCmd(gradleRunConfig); err != nil {
		return err
	}
	return nil
}

func downloadGradleDependencies() (gradleDependenciesDir, gradlePluginFilename string, err error) {
	dependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return
	}
	gradleDependenciesDir = filepath.Join(dependenciesPath, "gradle", gradleExtractorDependencyVersion)
	gradlePluginFilename = fmt.Sprintf("build-info-extractor-gradle-%s-uber.jar", gradleExtractorDependencyVersion)

	filePath := fmt.Sprintf("jfrog/jfrog-jars/org/jfrog/buildinfo/build-info-extractor-gradle/%s", gradleExtractorDependencyVersion)
	downloadPath := path.Join(filePath, gradlePluginFilename)

	err = utils.DownloadFromBintrayIfNeeded(downloadPath, gradlePluginFilename, gradleDependenciesDir)
	return
}

func createGradleRunConfig(tasks, configPath string, configuration *utils.BuildConfiguration, gradleDependenciesDir, gradlePluginFilename string) (*gradleRunConfig, error) {
	runConfig := &gradleRunConfig{env: map[string]string{}}
	runConfig.tasks = tasks

	vConfig, err := utils.ReadConfigFile(configPath, utils.YAML)
	if err != nil {
		return nil, err
	}

	runConfig.gradle, err = getGradleExecPath(vConfig.GetBool(useWrapper))
	if err != nil {
		return nil, err
	}

	runConfig.env[gradleBuildInfoProperties], err = utils.CreateBuildInfoPropertiesFile(configuration.BuildName, configuration.BuildNumber, vConfig, utils.GRADLE)
	if err != nil {
		return nil, err
	}

	if !vConfig.GetBool(usePlugin) {
		runConfig.initScript, err = getInitScript(gradleDependenciesDir, gradlePluginFilename)
		if err != nil {
			return nil, err
		}
	}

	return runConfig, nil
}

func getInitScript(gradleDependenciesDir, gradlePluginFilename string) (string, error) {
	gradleDependenciesDir, err := filepath.Abs(gradleDependenciesDir)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	initScriptPath := filepath.Join(gradleDependenciesDir, gradleInitScriptTemplate)

	exists, err := fileutils.IsFileExists(initScriptPath)
	if exists || err != nil {
		return initScriptPath, err
	}

	gradlePluginPath := filepath.Join(gradleDependenciesDir, gradlePluginFilename)
	gradlePluginPath = strings.Replace(gradlePluginPath, "\\", "\\\\", -1)
	initScriptContent := strings.Replace(utils.GradleInitScript, "${pluginLibDir}", gradlePluginPath, -1)
	if !fileutils.IsPathExists(gradleDependenciesDir) {
		err = os.MkdirAll(gradleDependenciesDir, 0777)
		if errorutils.CheckError(err) != nil {
			return "", err
		}
	}

	return initScriptPath, errorutils.CheckError(ioutil.WriteFile(initScriptPath, []byte(initScriptContent), 0644))
}

type gradleRunConfig struct {
	gradle     string
	tasks      string
	initScript string
	env        map[string]string
}

func (config *gradleRunConfig) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.gradle)
	if config.initScript != "" {
		cmd = append(cmd, "--init-script", config.initScript)
	}
	cmd = append(cmd, strings.Split(config.tasks, " ")...)

	log.Info("Running gradle command:", strings.Join(cmd, " "))
	return exec.Command(cmd[0], cmd[1:]...)
}

func (config *gradleRunConfig) GetEnv() map[string]string {
	return config.env
}

func (config *gradleRunConfig) GetStdWriter() io.WriteCloser {
	return nil
}

func (config *gradleRunConfig) GetErrWriter() io.WriteCloser {
	return nil
}

func getGradleExecPath(useWrapper bool) (string, error) {
	if useWrapper {
		if runtime.GOOS == "windows" {
			return "gradlew.bat", nil
		}
		return "./gradlew", nil
	}
	gradleExec, err := exec.LookPath("gradle")
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return gradleExec, nil
}
