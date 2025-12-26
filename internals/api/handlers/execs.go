package handlers

import (
	"ClassConnectRPC/internals/api/interceptors"
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/internals/repositories/mongodb"
	"ClassConnectRPC/pkg/utils"
	pb "ClassConnectRPC/proto/gen"
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) AddExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	for _, exec := range req.GetExecs() {
		if exec.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "Request is in incorrect format: non-empty ID field is not allowed")
		}
	}

	addedExecs, err := mongodb.AddExecsToDb(ctx, req.GetExecs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: addedExecs}, nil
}

func (s *Server) GetExecs(ctx context.Context, req *pb.GetExecsRequest) (*pb.Execs, error) {
	// We only allow people with admin and manager role to be able to use this endpoint
	// err := utils.AuthorizeUser(ctx, "admin", "manager")
	// if err != nil {
	// 	return nil, utils.ErrorHandler(err, err.Error())

	// }
	// Getting all the filters
	filters, err := buildFilterForModel(req.Exec, &models.Exec{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Getting all the sorting options
	sortOptions := buildSortOptions(req.SortBy)

	// Querying the database
	execs, err := mongodb.GetExecsFromDB(ctx, sortOptions, filters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Execs{Execs: execs}, nil

}

func (s *Server) UpdateExecs(ctx context.Context, req *pb.Execs) (*pb.Execs, error) {
	updatedExecs, err := mongodb.ModifyExecsInDB(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Execs{Execs: updatedExecs}, nil
}

func (s *Server) DeleteExecs(ctx context.Context, req *pb.ExecIds) (*pb.DeleteExecsConfirmation, error) {
	ids := req.GetIds()
	var objectIdsToDelete []primitive.ObjectID
	for _, v := range ids {
		if v.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "ID field cannot be empty")
		}
		objId, err := primitive.ObjectIDFromHex(v.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		objectIdsToDelete = append(objectIdsToDelete, objId)
	}

	deletedIds, err := mongodb.DeleteExecsFromDB(ctx, objectIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteExecsConfirmation{
		Status:     "Execs successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.ExecLoginRequest) (*pb.ExecLoginResponse, error) {

	exec, err := mongodb.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if exec.InactiveStatus {
		return nil, status.Error(codes.Unauthenticated, "Account is inactive")
	}

	err = utils.VerifyPassword(req.GetPassword(), exec.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Incorrect username or password")
	}

	tokenString, err := utils.SignToken(exec.Id, exec.Username, exec.Role)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Could not create token")
	}

	return &pb.ExecLoginResponse{Status: true, Token: tokenString}, nil
}

func (s *Server) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	exec, err := mongodb.UpdateUserInDB(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	token, err := utils.SignToken(exec.Id, exec.Username, exec.Role)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to generate token")
	}

	return &pb.UpdatePasswordResponse{
		PasswordUpdated: true,
		Token:           token,
	}, nil
}

func (s *Server) DeactivateUser(ctx context.Context, req *pb.ExecIds) (*pb.Confirmation, error) {
	objIds := []primitive.ObjectID{}

	for _, execId := range req.GetIds() {
		objId, err := primitive.ObjectIDFromHex(execId.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}
		objIds = append(objIds, objId)
	}

	result, err := mongodb.DeactivateUserInDB(ctx, objIds)
	if err != nil {
		return result, err
	}
	return &pb.Confirmation{Confirmation: true}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.EmptyRequest) (*pb.ExecLogoutResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "No metadata found")
	}

	val, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized access")
	}

	token := strings.TrimPrefix(val[0], "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized access")
	}

	// Get expiry time from context (set by authentication interceptor)
	expiryTimeStamp := ctx.Value(interceptors.ContextKey("expiresAt"))
	expiryTimeInt, ok := expiryTimeStamp.(int64)
	if !ok {
		return nil, status.Error(codes.Internal, "Failed to retrieve token expiry time")
	}

	expiryTime := time.Unix(expiryTimeInt, 0)

	utils.JwtStore.AddToken(token, expiryTime)

	return &pb.ExecLogoutResponse{
		LoggedOut: true,
	}, nil
}
