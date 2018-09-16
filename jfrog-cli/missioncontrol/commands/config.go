package commands

import (
	"errors"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/config"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/ioutils"
	"github.com/AlexeiVainshtein/jfrog-cli-go-modules/jfrog-cli/utils/lock"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/errorutils"
	"github.com/AlexeiVainshtein/jfrog-client-go/utils/log"
	"net/url"
	"sync"
)

// Internal golang locking for the same process.
var mutux sync.Mutex

func GetConfig() (*config.MissionControlDetails, error) {
	return config.ReadMissionControlConf()
}

func ShowConfig() error {
	details, err := config.ReadMissionControlConf()
	if err != nil {
		return err
	}
	if details.Url != "" {
		log.Output("Url: " + details.Url)
	}
	if details.User != "" {
		log.Output("User: " + details.User)
	}
	if details.Password != "" {
		log.Output("Password: ***")
	}
	return nil
}

func ClearConfig() {
	config.SaveMissionControlConf(new(config.MissionControlDetails))
}

func Config(details, defaultDetails *config.MissionControlDetails, interactive bool) (conf *config.MissionControlDetails, err error) {
	mutux.Lock()
	lockFile, err := lock.CreateLock()
	defer mutux.Unlock()
	defer lockFile.Unlock()

	if err != nil {
		return nil, err
	}

	allowUsingSavedPassword := true
	conf = details
	if conf == nil {
		conf = new(config.MissionControlDetails)
	}
	if interactive {
		if defaultDetails == nil {
			defaultDetails, err = config.ReadMissionControlConf()
			if err != nil {
				return
			}
		}
		if conf.Url == "" {
			ioutils.ScanFromConsole("Mission Control URL", &conf.Url, defaultDetails.Url)
			var u *url.URL
			u, err = url.Parse(conf.Url)
			err = errorutils.CheckError(err)
			if err != nil {
				return
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				err = errorutils.CheckError(errors.New("URL scheme is not valid " + u.Scheme))
				if err != nil {
					return
				}
			}
			allowUsingSavedPassword = false
		}
		ioutils.ReadCredentialsFromConsole(conf, defaultDetails, allowUsingSavedPassword)
	}
	conf.Url = utils.AddTrailingSlashIfNeeded(conf.Url)
	config.SaveMissionControlConf(conf)
	return
}

type ConfigFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive           bool
}
