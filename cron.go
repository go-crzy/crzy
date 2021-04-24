package crzy

import (
	"context"
	"log"

	cron "github.com/robfig/cron/v3"
)

type CronService struct {
	cron *cron.Cron
}

func NewCronService() *CronService {
	cron := cron.New()
	return &CronService{
		cron: cron,
	}
}

func (c *CronService) Run(ctx context.Context) error {
	log.Println("starting cron....")
	c.cron.Start()
	<-ctx.Done()
	log.Println("stopping cron....")
	c.cron.Stop()
	return ctx.Err()
}
