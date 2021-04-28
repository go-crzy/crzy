package crzy

import (
	"context"
	"log"
	"os"
)

type StoreService struct {
}

func NewStoreService() *StoreService {
	return &StoreService{}
}

func (m *StoreService) Run(ctx context.Context) error {
	log.Println("starting datastore....")
	<-ctx.Done()
	log.Println("stopping datastore....")
	err := os.RemoveAll("/tmp/workspace")
	if err != nil {
		log.Printf("error deleting /tmp/workspace: %v", err)
	}
	log.Println("datastore stopped....")
	return ctx.Err()
}
