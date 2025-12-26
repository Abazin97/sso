package auth

import (
	"context"
	"errors"
	"sso/internal/domain/models"
	"sso/internal/repository"
	"sso/internal/services/auth"

	ssov1 "github.com/Abazin97/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		phone string,
		appID int,
	) (models.User, string, error)
	RegisterNewUser(ctx context.Context,
		title string,
		birthDate string,
		name string,
		lastname string,
		email string,
		password string,
		phone string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	ChangePasswordInit(
		ctx context.Context,
		email string,
		phone string,
		oldPassword string) (string, int64, error)
	ChangePasswordConfirm(
		ctx context.Context,
		code string,
		verificationID int64,
		email string,
		newPassword string,
	) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	if err := validateLogin(req); err != nil {
		return nil, err
	}

	user, token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetPhone(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		if errors.Is(err, repository.ErrAppNotFound) {
			return nil, status.Error(codes.InvalidArgument, "invalid app id")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov1.LoginResponse{
		User:  ToProtoUser(user),
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	// TODO: auth per OTP code below
	//resp, err := s.auth.RequestOTP(req.GetPhone())
	//if err != nil {
	//	return nil, status.Error(codes.Internal, "failed to request OTP")
	//}
	//
	//err = s.auth.VerifyOTP(req.GetPhone(), resp)
	//if err != nil {
	//	return nil, status.Error(codes.Internal, "failed to verify OTP")
	//}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetTitle(), req.GetBirthDate(), req.GetName(), req.GetLastName(), req.GetEmail(), req.GetPassword(), req.GetPhone())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) ChangePasswordInit(
	ctx context.Context,
	req *ssov1.ChangePassInitRequest) (*ssov1.ChangePassInitResponse, error) {

	expTime, verID, err := s.auth.ChangePasswordInit(ctx, req.GetEmail(), req.GetPhone(), req.GetOldPassword())

	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &ssov1.ChangePassInitResponse{ExpiryTime: expTime, Uid: verID}, nil
}

func (s *serverAPI) ChangePasswordConfirm(
	ctx context.Context,
	req *ssov1.ChangePassConfirmRequest) (*ssov1.ChangePassConfirmResponse, error) {

	success, err := s.auth.ChangePasswordConfirm(ctx, req.GetCode(), req.GetUid(), req.GetEmail(), req.GetNewPassword())

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &ssov1.ChangePassConfirmResponse{Success: success}, nil

}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetPhone() == "" {
		return status.Error(codes.InvalidArgument, "phone is required")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	return nil
}
