package pkg

import (
	"context"

	"github.com/go-logr/logr"
	cron "github.com/robfig/cron/v3"
)

type CronService struct {
	cron *cron.Cron
	log  logr.Logger
}

func NewCronService() *CronService {
	cron := cron.New()
	return &CronService{
		cron: cron,
		log:  NewLogger("cron"),
	}
}

func (c *CronService) Run(ctx context.Context) error {
	log := c.log
	log.Info("starting cron....")
	c.cron.Start()
	<-ctx.Done()
	log.Info("stopping cron....")
	c.cron.Stop()
	return ctx.Err()
}
