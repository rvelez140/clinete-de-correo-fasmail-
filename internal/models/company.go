package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Company struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	LogoPath     string    `json:"logo_path"`
	PrimaryColor string    `json:"primary_color"`
	SuccessColor string    `json:"success_color"`
	DangerColor  string    `json:"danger_color"`
	WarningColor string    `json:"warning_color"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CompanyBranding struct {
	Name         string
	LogoURL      string
	PrimaryColor string
	SuccessColor string
	DangerColor  string
	WarningColor string
}

func (c *Company) ToBranding(uploadsBase string) CompanyBranding {
	logoURL := ""
	if c.LogoPath != "" {
		logoURL = uploadsBase + "/" + c.LogoPath
	}
	return CompanyBranding{
		Name:         c.Name,
		LogoURL:      logoURL,
		PrimaryColor: c.PrimaryColor,
		SuccessColor: c.SuccessColor,
		DangerColor:  c.DangerColor,
		WarningColor: c.WarningColor,
	}
}

type CompanyRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{pool: pool}
}

func (r *CompanyRepository) Create(ctx context.Context, company *Company) error {
	query := `
		INSERT INTO companies (name, slug, logo_path, primary_color, success_color, danger_color, warning_color, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		company.Name,
		company.Slug,
		company.LogoPath,
		company.PrimaryColor,
		company.SuccessColor,
		company.DangerColor,
		company.WarningColor,
		company.IsActive,
	).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
}

func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*Company, error) {
	query := `
		SELECT id, name, slug, logo_path, primary_color, success_color, danger_color, warning_color,
		       is_active, created_at, updated_at
		FROM companies WHERE id = $1`

	company := &Company{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&company.ID, &company.Name, &company.Slug, &company.LogoPath,
		&company.PrimaryColor, &company.SuccessColor, &company.DangerColor, &company.WarningColor,
		&company.IsActive, &company.CreatedAt, &company.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("empresa no encontrada")
		}
		return nil, fmt.Errorf("get company: %w", err)
	}

	return company, nil
}

func (r *CompanyRepository) GetBySlug(ctx context.Context, slug string) (*Company, error) {
	query := `
		SELECT id, name, slug, logo_path, primary_color, success_color, danger_color, warning_color,
		       is_active, created_at, updated_at
		FROM companies WHERE slug = $1 AND is_active = true`

	company := &Company{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&company.ID, &company.Name, &company.Slug, &company.LogoPath,
		&company.PrimaryColor, &company.SuccessColor, &company.DangerColor, &company.WarningColor,
		&company.IsActive, &company.CreatedAt, &company.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("empresa no encontrada")
		}
		return nil, fmt.Errorf("get company by slug: %w", err)
	}

	return company, nil
}

func (r *CompanyRepository) GetDefault(ctx context.Context) (*Company, error) {
	return r.GetBySlug(ctx, "default")
}

func (r *CompanyRepository) Update(ctx context.Context, company *Company) error {
	query := `
		UPDATE companies SET name = $1, slug = $2, logo_path = $3,
		       primary_color = $4, success_color = $5, danger_color = $6, warning_color = $7,
		       is_active = $8, updated_at = NOW()
		WHERE id = $9`

	tag, err := r.pool.Exec(ctx, query,
		company.Name,
		company.Slug,
		company.LogoPath,
		company.PrimaryColor,
		company.SuccessColor,
		company.DangerColor,
		company.WarningColor,
		company.IsActive,
		company.ID,
	)
	if err != nil {
		return fmt.Errorf("update company: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("empresa no encontrada")
	}

	return nil
}

func (r *CompanyRepository) List(ctx context.Context, offset, limit int) ([]Company, int, error) {
	countQuery := `SELECT COUNT(*) FROM companies`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count companies: %w", err)
	}

	query := `
		SELECT id, name, slug, logo_path, primary_color, success_color, danger_color, warning_color,
		       is_active, created_at, updated_at
		FROM companies ORDER BY created_at ASC LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Slug, &c.LogoPath,
			&c.PrimaryColor, &c.SuccessColor, &c.DangerColor, &c.WarningColor,
			&c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan company: %w", err)
		}
		companies = append(companies, c)
	}

	return companies, total, nil
}

func (r *CompanyRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM companies`).Scan(&count)
	return count, err
}

func (r *CompanyRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM companies WHERE slug = $1)`, slug).Scan(&exists)
	return exists, err
}

func (r *CompanyRepository) UpdateLogoPath(ctx context.Context, id uuid.UUID, logoPath string) error {
	query := `UPDATE companies SET logo_path = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, logoPath, id)
	if err != nil {
		return fmt.Errorf("update logo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("empresa no encontrada")
	}
	return nil
}
