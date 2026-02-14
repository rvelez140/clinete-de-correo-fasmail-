package admin

import (
	"context"
	"fmt"
	"regexp"

	"github.com/fasmail/panel/internal/models"
	"github.com/google/uuid"
)

var (
	colorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	slugRegex  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
)

type CompanyService struct {
	companyRepo *models.CompanyRepository
	userRepo    *models.UserRepository
}

func NewCompanyService(companyRepo *models.CompanyRepository, userRepo *models.UserRepository) *CompanyService {
	return &CompanyService{companyRepo: companyRepo, userRepo: userRepo}
}

func (s *CompanyService) CreateCompany(ctx context.Context, name, slug, primaryColor, successColor, dangerColor, warningColor string) (*models.Company, error) {
	if name == "" {
		return nil, fmt.Errorf("el nombre es requerido")
	}

	if len(slug) < 2 {
		return nil, fmt.Errorf("el slug debe tener al menos 2 caracteres")
	}

	if !slugRegex.MatchString(slug) {
		return nil, fmt.Errorf("el slug solo puede contener letras minúsculas, números y guiones")
	}

	exists, err := s.companyRepo.SlugExists(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("verificar slug: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("el slug '%s' ya está en uso", slug)
	}

	if err := validateColor(primaryColor); err != nil {
		return nil, fmt.Errorf("color primario: %w", err)
	}
	if err := validateColor(successColor); err != nil {
		return nil, fmt.Errorf("color éxito: %w", err)
	}
	if err := validateColor(dangerColor); err != nil {
		return nil, fmt.Errorf("color peligro: %w", err)
	}
	if err := validateColor(warningColor); err != nil {
		return nil, fmt.Errorf("color advertencia: %w", err)
	}

	company := &models.Company{
		Name:         name,
		Slug:         slug,
		PrimaryColor: primaryColor,
		SuccessColor: successColor,
		DangerColor:  dangerColor,
		WarningColor: warningColor,
		IsActive:     true,
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, fmt.Errorf("crear empresa: %w", err)
	}

	return company, nil
}

func (s *CompanyService) UpdateCompany(ctx context.Context, company *models.Company) error {
	if company.Name == "" {
		return fmt.Errorf("el nombre es requerido")
	}

	if len(company.Slug) < 2 {
		return fmt.Errorf("el slug debe tener al menos 2 caracteres")
	}

	if !slugRegex.MatchString(company.Slug) {
		return fmt.Errorf("el slug solo puede contener letras minúsculas, números y guiones")
	}

	if err := validateColor(company.PrimaryColor); err != nil {
		return fmt.Errorf("color primario: %w", err)
	}
	if err := validateColor(company.SuccessColor); err != nil {
		return fmt.Errorf("color éxito: %w", err)
	}
	if err := validateColor(company.DangerColor); err != nil {
		return fmt.Errorf("color peligro: %w", err)
	}
	if err := validateColor(company.WarningColor); err != nil {
		return fmt.Errorf("color advertencia: %w", err)
	}

	return s.companyRepo.Update(ctx, company)
}

func (s *CompanyService) UpdateCompanyLogo(ctx context.Context, companyID uuid.UUID, logoPath string) error {
	return s.companyRepo.UpdateLogoPath(ctx, companyID, logoPath)
}

func (s *CompanyService) DeleteCompany(ctx context.Context, companyID uuid.UUID) error {
	company, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	if company.Slug == "default" {
		return fmt.Errorf("no se puede eliminar la empresa por defecto")
	}

	company.IsActive = false
	return s.companyRepo.Update(ctx, company)
}

func (s *CompanyService) ListCompanies(ctx context.Context, offset, limit int) ([]models.Company, int, error) {
	return s.companyRepo.List(ctx, offset, limit)
}

func (s *CompanyService) GetCompany(ctx context.Context, companyID uuid.UUID) (*models.Company, error) {
	return s.companyRepo.GetByID(ctx, companyID)
}

func validateColor(color string) error {
	if !colorRegex.MatchString(color) {
		return fmt.Errorf("formato de color inválido (use #RRGGBB)")
	}
	return nil
}
