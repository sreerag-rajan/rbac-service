package repository

import (
	"context"
	"fmt"
	"rbac-service/internal/model"
)

type ResourceRepository struct{}

func NewResourceRepository() *ResourceRepository {
	return &ResourceRepository{}
}

func (r *ResourceRepository) GetResourceByCode(ctx context.Context, code string) (*model.Resource, error) {
	pool := GetPool()
	var res model.Resource
	err := pool.QueryRow(ctx, "SELECT id, code, name, description FROM pmsn.resource WHERE code = $1", code).Scan(&res.ID, &res.Code, &res.Name, &res.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource by code: %w", err)
	}
	return &res, nil
}

func (r *ResourceRepository) GetActionByCode(ctx context.Context, resourceID, actionCode string) (*model.Action, error) {
	pool := GetPool()
	var act model.Action
	err := pool.QueryRow(ctx, "SELECT id, resource_id, code, name, description FROM pmsn.action WHERE resource_id = $1 AND code = $2", resourceID, actionCode).Scan(&act.ID, &act.ResourceID, &act.Code, &act.Name, &act.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to get action by code: %w", err)
	}
	return &act, nil
}
