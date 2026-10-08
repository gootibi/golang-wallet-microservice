package service

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gootibi/golang-wallet-microservice/monolith/internal/auth"
	customError "github.com/gootibi/golang-wallet-microservice/monolith/internal/errors"
	userModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/model"
	userRepisitory "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
	walletModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/model"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req userModel.CreateUserRequest) (*userModel.User, error)
	GetProfile(ctx context.Context, id string) (*userModel.User, error)
	UpdateProfile(ctx context.Context, id string, req userModel.UpdateUserRequest) (*userModel.User, error)
	Login(ctx context.Context, req userModel.LoginRequest) (*userModel.LoginResponse, error)
}

func NewuserService(db *sql.DB, uRepo userRepisitory.UserRepository, wRepo walletRepository.WalletRepository) UserService {
	return &userService{
		db:         db,
		userRepo:   uRepo,
		walletRepo: wRepo,
	}
}

type userService struct {
	db         *sql.DB
	userRepo   userRepisitory.UserRepository
	walletRepo walletRepository.WalletRepository
}

// Register implements [UserService].
func (s *userService) Register(ctx context.Context, req userModel.CreateUserRequest) (*userModel.User, error) {
	// 1. Check if the email is already register
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		// Return custom AppError
		return nil, customError.NewAppError(http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "This email already registered.")
	}

	// Hash the pasword with bcrypt
	hasedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		// Return internal server error
		return nil, customError.ErrInternalServer
	}

	// 2. Create new user object
	user := &userModel.User{
		ID:           uuid.New().String(),
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hasedBytes),
	}

	// Begin transaction database
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	// We should rollback if anything error or panic in the middle
	defer tx.Rollback()

	// Store user to db with a tx connection
	if err := s.userRepo.CreateTx(ctx, tx, user); err != nil {
		return nil, customError.ErrInternalServer
	}

	// Create Wallet for the user
	wallet := &walletModel.Wallet{
		ID:       uuid.New().String(),
		UserID:   user.ID,
		Balance:  0.0,
		Currency: "IDR",
		Status:   "active",
	}
	if err := s.walletRepo.CreateTx(ctx, tx, wallet); err != nil {
		return nil, customError.ErrInternalServer
	}

	// Commit the transaction if all of the step is success
	if err := tx.Commit(); err != nil {
		return nil, customError.ErrInternalServer
	}

	return s.userRepo.GetByID(ctx, user.ID)
}

// GetProfile implements [UserService].
func (s *userService) GetProfile(ctx context.Context, id string) (*userModel.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	}

	return u, nil
}

// UpdateProfile implements [UserService].
func (s *userService) UpdateProfile(ctx context.Context, id string, req userModel.UpdateUserRequest) (*userModel.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, customError.NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	}

	user.FullName = req.FullName
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, customError.ErrInternalServer
	}

	return s.userRepo.GetByID(ctx, id)
}

// Login implements [UserService].
func (s *userService) Login(ctx context.Context, req userModel.LoginRequest) (*userModel.LoginResponse, error) {
	// Find by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, customError.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Wrong email or password")
	}

	// Verify the hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, customError.NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Wrong email or password")
	}

	// Generate access token 15 minutes
	accessToken, err := auth.GenerateToken(user.ID, user.Email, 15*time.Minute)
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	// Generate refresh token 7 days
	refreshToken, err := auth.GenerateToken(user.ID, user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, customError.ErrInternalServer
	}

	// Return the tokens
	return &userModel.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
