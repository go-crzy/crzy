package pkg

import (
	"context"
	"os"

	"github.com/go-logr/logr"
)

type StoreService struct {
	workspace string
	log       logr.Logger
}

func NewStoreService(workspace string) *StoreService {
	return &StoreService{
		workspace: workspace,
		log:       NewLogger("store"),
	}
}

func (m *StoreService) Run(ctx context.Context) error {
	log := m.log
	log.Info("starting datastore....")
	<-ctx.Done()
	log.Info("stopping datastore....")
	err := os.RemoveAll(m.workspace)
	if err != nil {
		log.Error(err, "error deleting workspace", "data", m.workspace)
	}
	log.Info("datastore stopped....")
	return ctx.Err()
}
