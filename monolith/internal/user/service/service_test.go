package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	userModel "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/model"
	userRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/user/repository"
	walletRepository "github.com/gootibi/golang-wallet-microservice/monolith/internal/wallet/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
	}

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, "john.doe@example.com").Return(nil, errors.New("User not found"))
	mockUserRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)
	mockWalletRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)

	expectUser := &userModel.User{
		ID:       "some-uuid",
		FullName: req.FullName,
		Email:    req.Email,
	}
	mockUserRepo.On("GetByID", ctx, mock.Anything).Return(expectUser, nil)

	// Run the function
	user, err := svc.Register(ctx, req)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.FullName, user.FullName)
	assert.Equal(t, req.Email, user.Email)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	// Sql mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
	}

	// Email already in db
	existingUser := &userModel.User{
		FullName: req.FullName,
		Email:    req.Email,
	}

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	// Run the function
	user, err := svc.Register(ctx, req)

	// Verify result should be error and no return the user object
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "This email already registered.", err.Error())

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestRegister_UserCreateFailed(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
	}

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, "john.doe@example.com").Return(nil, errors.New("User not found"))
	mockUserRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(errors.New("db error"))

	// Run the function
	user, err := svc.Register(ctx, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, user)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestRegister_WalletCreateFailed(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
	}

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, "john.doe@example.com").Return(nil, errors.New("User not found"))
	mockUserRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)
	mockWalletRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(errors.New("db error"))

	// Run the function
	user, err := svc.Register(ctx, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, user)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	assert.Error(t, dbMock.ExpectationsWereMet())
}

func TestRegister_TxCommitError(t *testing.T) {
	// Sql mock
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.CreateUserRequest{
		FullName: "John Doe",
		Email:    "john.doe@example.com",
		Password: "password",
	}

	// Expect sql transaction
	dbMock.ExpectBegin()
	dbMock.ExpectCommit().WillReturnError(errors.New("Internal server errror"))

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, "john.doe@example.com").Return(nil, errors.New("User not found"))
	mockUserRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)
	mockWalletRepo.On("CreateTx", ctx, mock.Anything, mock.Anything).Return(nil)

	expectUser := &userModel.User{
		ID:       "some-uuid",
		FullName: req.FullName,
		Email:    req.Email,
	}
	mockUserRepo.On("GetByID", ctx, mock.Anything).Return(expectUser, nil)

	// Run the function
	user, err := svc.Register(ctx, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, user)

	// Make sure all mock if called and expected
	mockWalletRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGetProfile_Seccess(t *testing.T) {
	// Sql mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "user-1234"
	expectUser := &userModel.User{
		ID:       userID,
		FullName: "John Doe",
		Email:    "john.doe@example.com",
	}

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(expectUser, nil)

	// Run the function
	user, err := svc.GetProfile(ctx, userID)

	// Verify result
	assert.NoError(t, err)
	assert.Equal(t, expectUser, user)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	// Sql mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "non-exist"

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("User not found"))

	// Run the function
	user, err := svc.GetProfile(ctx, userID)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, user)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestUpdateProfile_Success(t *testing.T) {
	// Sql mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"
	req := userModel.UpdateUserRequest{
		FullName: "Todor Doe",
	}

	existingUser := &userModel.User{
		ID:       userID,
		FullName: "John Doe",
		Email:    "john.doe@example.com",
	}
	updateUser := &userModel.User{
		ID:       userID,
		FullName: "Todor Doe",
		Email:    "john.doe@example.com",
	}

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
	mockUserRepo.On("Update", ctx, mock.Anything).Return(nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(updateUser, nil).Once()

	// Run the function
	user, err := svc.UpdateProfile(ctx, userID, req)

	// Verify result
	assert.NoError(t, err)
	assert.Equal(t, "Todor Doe", user.FullName)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestUpdateProfile_NotFound(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "non-existing"
	req := userModel.UpdateUserRequest{
		FullName: "Todor Doe",
	}

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("User not found"))

	// Run the function
	user, err := svc.UpdateProfile(ctx, userID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, user)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestUpdateProfile_UpdateFailure(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"
	req := userModel.UpdateUserRequest{
		FullName: "Todor Doe",
	}
	exisitingUser := &userModel.User{
		ID:       userID,
		FullName: "John Doe",
		Email:    "john.doe@example.com",
	}

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(exisitingUser, nil)
	mockUserRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	// Run the function
	user, err := svc.UpdateProfile(ctx, userID, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, user)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "password",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	existingUser := &userModel.User{
		ID:           "some_uuid",
		FullName:     "John Doe",
		Email:        "john.doe@example.com",
		PasswordHash: string(hashedPassword),
	}

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	// Run the function
	response, err := svc.Login(ctx, req)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials_EmailNotFound(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "password",
	}

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("Wrong email or password"))

	// Run the function
	response, err := svc.Login(ctx, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, response)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials_WrongPassword(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	req := userModel.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "wrong-password",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	exisitingUser := &userModel.User{
		ID:           "some-uuid",
		FullName:     "John Doe",
		Email:        "john.doe@example.com",
		PasswordHash: string(hashedPassword),
	}

	// Expect mock
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(exisitingUser, nil)

	// Run the function
	response, err := svc.Login(ctx, req)

	// Verify result
	assert.Error(t, err)
	assert.Nil(t, response)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestUpdateAvatar_Success(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"
	path := "/uploads/avatar.png"

	mockUserRepo.On("UpdateAvatar", ctx, userID, path).Return(nil)

	// Run the function
	err := svc.UpdateAvatar(ctx, userID, path)

	// Verify result
	assert.NoError(t, err)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestUpdateAvatar_Failure(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userId := "some-uuid"
	path := "/uploads/avatar.png"

	// Expect mock
	mockUserRepo.On("UpdateAvatar", ctx, userId, path).Return(errors.New("db error"))

	// Run the function
	err := svc.UpdateAvatar(ctx, userId, path)

	// Verify result
	assert.Error(t, err)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestDeleteAccount_Success(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"
	existingUser := &userModel.User{
		ID:       userID,
		FullName: "John Doe",
		Email:    "john.doe@example.com",
	}

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockUserRepo.On("SoftDelete", ctx, userID).Return(nil)

	// Run the function
	err := svc.DeleteAccount(ctx, userID)

	// Verify result
	assert.NoError(t, err)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestDeleteAccount_NotFound(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("User not found"))

	// Run the function
	err := svc.DeleteAccount(ctx, userID)

	// Verify result
	assert.Error(t, err)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestDeleteAccount_SoftDeleteFailure(t *testing.T) {
	// Sql Mock
	db, _, _ := sqlmock.New()
	defer db.Close()

	// Initiate mock reposirories
	mockUserRepo := new(userRepository.MockUserRepository)
	mockWalletRepo := new(walletRepository.MockWalletRepository)

	// We no need real sql.DB connection, so we sent sql mock
	svc := NewUserService(db, mockUserRepo, mockWalletRepo)

	ctx := context.TODO()
	userID := "some-uuid"
	existingUser := &userModel.User{
		ID:       userID,
		FullName: "John Doe",
		Email:    "john.doe@example.com",
	}

	// Expect mock
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockUserRepo.On("SoftDelete", ctx, userID).Return(errors.New("db error"))

	// Run the function
	err := svc.DeleteAccount(ctx, userID)

	// Verify result
	assert.Error(t, err)

	// Make sure all mock if called and expected
	mockUserRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}
