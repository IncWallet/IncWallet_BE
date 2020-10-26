package jobs

import (
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
	"incwallet/app/controllers"
	"time"
)

type UpdateCommitteeJob struct {
	Name string
}
type UpdateCommitteeRewardJob struct {
	Name string
}

func (j UpdateCommitteeJob)Run()  {
	revel.AppLog.Infof("%s %s", j.Name, time.Now().String())
	err := controllers.StateM.CommitteManager.UpdateCommitteeFromChain()
	if err != nil{
		revel.AppLog.Warnf("cannot update committee from job. Error %v", err)
	}
}

func (j UpdateCommitteeRewardJob)Run()  {
	revel.AppLog.Infof("%s %s", j.Name, time.Now().String())
	err := controllers.StateM.CommitteManager.UpdateMinerRewardFromChain()
	if err != nil{
		revel.AppLog.Warnf("cannot update reward from job. Error %v", err)
	}
}

func init()  {
	revel.OnAppStart(func() {
		jobs.Every(5*time.Minute, UpdateCommitteeJob{Name: "Update committee job now..."})
		jobs.Every(5*time.Minute, UpdateCommitteeRewardJob{Name: "Update reward job now..."})
	})
}
