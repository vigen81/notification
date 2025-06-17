package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"go.uber.org/zap"
)

type TemplateService struct {
	templateRepo    *repository.TemplateRepository
	localizationSvc *LocalizationService
	logger          *zap.Logger
}

func NewTemplateService(
	templateRepo *repository.TemplateRepository,
	localizationSvc *LocalizationService,
	logger *zap.Logger,
) *TemplateService {
	return &TemplateService{
		templateRepo:    templateRepo,
		localizationSvc: localizationSvc,
		logger:          logger,
	}
}

func (s *TemplateService) GetTemplate(ctx context.Context, tenantID int64, templateID string) (*models.Template, error) {
	return s.templateRepo.GetByID(ctx, tenantID, templateID)
}

func (s *TemplateService) ProcessTemplate(ctx context.Context, tmpl *models.Template, data map[string]interface{}, locale string) (body string, headline string, err error) {
	// Apply localization
	localizedData := s.localizationSvc.LocalizeData(locale, data)

	// Process body template
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse body template: %w", err)
	}

	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, localizedData); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}

	// Process headline template if exists
	if tmpl.Subject != "" {
		headlineTmpl, err := template.New("headline").Parse(tmpl.Subject)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse headline template: %w", err)
		}

		var headlineBuf bytes.Buffer
		if err := headlineTmpl.Execute(&headlineBuf, localizedData); err != nil {
			return "", "", fmt.Errorf("failed to execute headline template: %w", err)
		}
		headline = headlineBuf.String()
	}

	return bodyBuf.String(), headline, nil
}

func (s *TemplateService) CreateTemplate(ctx context.Context, tenantID int64, req *models.TemplateRequest) (*models.Template, error) {
	// Validate template syntax
	if _, err := template.New("body").Parse(req.Body); err != nil {
		return nil, fmt.Errorf("invalid body template syntax: %w", err)
	}

	if req.Subject != "" {
		if _, err := template.New("subject").Parse(req.Subject); err != nil {
			return nil, fmt.Errorf("invalid subject template syntax: %w", err)
		}
	}

	return s.templateRepo.Create(ctx, tenantID, req)
}
