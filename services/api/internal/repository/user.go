package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leksa/datamapper-senyar/internal/model"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByOIDCSubject finds a user by OIDC subject (Authentik user ID)
func (r *UserRepository) FindByOIDCSubject(ctx context.Context, subject string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("oidc_subject = ?", subject).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.FindByID(ctx, uid)
}

func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

// FindWithOrganizations finds a user with their organization memberships
func (r *UserRepository) FindWithOrganizations(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("OrganizationMemberships.Organization").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserFilter defines filter options for listing users
type UserFilter struct {
	Role     string
	IsActive *bool
	Search   string
	Page     int
	Limit    int
}

func (r *UserRepository) FindByIDStr(ctx context.Context, id string) (*model.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, uid)
}

func (r *UserRepository) FindByInvitationToken(ctx context.Context, token string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("invitation_token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", uid).Error
}

func (r *UserRepository) FindByVerificationPIN(ctx context.Context, pin string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("verification_pin = ?", pin).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(ctx context.Context, filter UserFilter) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	// Apply filters
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) FindByVerificationPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("OrganizationMemberships").
		Where("verification_phone = ? AND status = ? AND verified_at IS NOT NULL", phone, model.UserStatusActive).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
