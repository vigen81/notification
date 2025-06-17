package repository

import (
	"context"

	"gitlab.smartbet.am/golang/notification/ent"
	//"gitlab.smartbet.am/golang/notification/ent/template"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"go.uber.org/zap"
)

type TemplateRepository struct {
	client *ent.Client
	logger *zap.Logger
}

func NewTemplateRepository(client *ent.Client, logger *zap.Logger) *TemplateRepository {
	return &TemplateRepository{
		client: client,
		logger: logger,
	}
}

func (r *TemplateRepository) GetByID(ctx context.Context, tenantID int64, templateID string) (*models.Template, error) {
	tmpl, err := r.client.Template.Query().
		Where(
			template.TenantID(tenantID),
			template.ID(templateID),
		).
		First(ctx)

	if err != nil {
		return nil, err
	}

	return r.entToModel(tmpl), nil
}

func (r *TemplateRepository) ListByTenant(ctx context.Context, tenantID int64) ([]*models.Template, error) {
	templates, err := r.client.Template.Query().
		Where(template.TenantID(tenantID)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.Template, len(templates))
	for i, tmpl := range templates {
		result[i] = r.entToModel(tmpl)
	}

	return result, nil
}

func (r *TemplateRepository) Create(ctx context.Context, tenantID int64, req *models.TemplateRequest) (*models.Template, error) {
	create := r.client.Template.Create().
		SetTenantID(tenantID).
		SetName(req.Name).
		SetType(template.Type(req.Type)).
		SetBody(req.Body)

	if req.Subject != "" {
		create.SetSubject(req.Subject)
	}

	if req.Metadata != nil {
		create.SetMetadata(req.Metadata)
	}

	tmpl, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.entToModel(tmpl), nil
}

func (r *TemplateRepository) entToModel(tmpl *ent.Template) *models.Template {
	model := &models.Template{
		ID:        tmpl.ID,
		TenantID:  tmpl.TenantID,
		Name:      tmpl.Name,
		Type:      models.NotificationType(tmpl.Type),
		Body:      tmpl.Body,
		CreatedAt: tmpl.CreateTime,
		UpdatedAt: tmpl.UpdateTime,
	}

	if tmpl.Subject != nil {
		model.Subject = *tmpl.Subject
	}

	if tmpl.Metadata != nil {
		model.Metadata = tmpl.Metadata
	}

	return model
}
