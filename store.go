package crzy

import (
	"context"
	"log"
	"os"
)

type StoreService struct {
	workspace string
}

func NewStoreService(workspace string) *StoreService {
	return &StoreService{
		workspace: workspace,
	}
}

func (m *StoreService) Run(ctx context.Context) error {
	log.Println("starting datastore....")
	<-ctx.Done()
	log.Println("stopping datastore....")
	err := os.RemoveAll(m.workspace)
	if err != nil {
		log.Printf("error deleting %s: %v", m.workspace, err)
	}
	log.Println("datastore stopped....")
	return ctx.Err()
}
