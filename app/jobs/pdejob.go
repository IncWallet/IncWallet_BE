package jobs

import (
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/controllers"
	"incwallet/app/database"
	"time"
)

type UpdatePdePoolPairsJob struct {
	Name string
}

type UpdatePdeHistoryJob struct {
	Name string
}

type UpdateXPdeHistoryJob struct {
	Name string
}

type CleanPdeDataJob struct {
	Name string
}

func (j UpdatePdePoolPairsJob) Run() {
	revel.AppLog.Infof("%s %s", j.Name, time.Now().String())

	err := controllers.StateM.PdeManager.UpdatePdePoolPairsFromChain()
	if err != nil {
		revel.AppLog.Warnf("cannot update pde pool pairs from job. Error %v", err)
	}
}

func (j UpdatePdeHistoryJob) Run() {
	revel.AppLog.Infof("%s %s", j.Name, time.Now().String())

	err := controllers.StateM.PdeManager.UpdatePdeTradeHistoryFromChain()
	if err != nil {
		revel.AppLog.Warnf("cannot update pde history from job. Error %v", err)
	}
}

func (j CleanPdeDataJob) Run() {
	revel.AppLog.Infof("%s %s", j.Name, time.Now().String())

	query := bson.M{
		"beaconheight": bson.M{
			"$lt": controllers.StateM.NetworkManager.BeaconState.Height,
		},
	}

	info, err := database.PoolPairs.RemoveAll(query)
	if err != nil {
		revel.AppLog.Warnf("cannot clean pde pool pairs from job. Error %v", err)
	}
	revel.AppLog.Infof("clean pde pool pairs. remove %v record(s)", info.Removed)}

//only run for backend service
func init() {
	revel.OnAppStart(func() {
		jobs.Every(40*time.Second, UpdatePdeHistoryJob{Name: "Update pde history job now..."})
		jobs.Every(40*time.Second, UpdatePdePoolPairsJob{Name: "Update pde pool pairs job now..."})
		jobs.Every(1*time.Hour, CleanPdeDataJob{"clean pde pool pairs job now..."})
	})
}
