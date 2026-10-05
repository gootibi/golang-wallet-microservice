package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/user/model"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
)

type UserService interface {
	Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error)
}

func NewuserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

type userService struct {
	repo repository.UserRepository
}

// Register implements [UserService].
func (s *userService) Register(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	// 1. Check if the email is already register
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// 2. Create new user object
	user := &model.User{
		ID:           uuid.New().String(),
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: req.Password, // TODO: hashing the password before storing it in the database
	}

	// 3. Store in the database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, user.ID)
}

// GetProfile implements [UserService].
func (s *userService) GetProfile(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateProfile implements [UserService].
func (s *userService) UpdateProfile(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FullName = req.FullName
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}
