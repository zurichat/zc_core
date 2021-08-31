package organizations

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type OrgService interface {
	Create(ctx context.Context, org Organization) (*mongo.InsertOneResult, error)
}

type Service struct {
	repo OrgRepository
}

func NewOrgService(a OrgRepository) OrgService {
	return &Service{
		repo: a,
	}
}

func (a *Service) Create(c context.Context, org Organization) (*mongo.InsertOneResult, error) {
	return a.repo.Create(c, org)	
}