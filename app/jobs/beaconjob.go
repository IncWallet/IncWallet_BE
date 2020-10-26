package jobs

import (
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"incwallet/app/controllers"
	"time"
)

type UpdateBeaconBestState struct {
	Name string
}

func (j UpdateBeaconBestState) Run() {
	revel.AppLog.Infof("%s %s", j.Name, time.Now().String())

	err := controllers.StateM.NetworkManager.UpdateBeaconState()
	if err != nil {
		revel.AppLog.Warnf("cannot update beacon best state. Error %v", err)
	}
}

//only run for backend service
func init() {
	revel.OnAppStart(func() {
		jobs.Every(20*time.Second, UpdateBeaconBestState{Name: "Update beacon best state job now..."})
	})
}
