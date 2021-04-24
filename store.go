package crzy

import (
	"context"
	"log"
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
	return ctx.Err()
}
