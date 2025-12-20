package handlers

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/internals/repositories/mongodb"
	pb "ClassConnectRPC/proto/gen"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	for _, teacher := range req.GetTeachers() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "Request is in incorrect format: non-empty ID field is not allowed")
		}
	}

	addedTeachers, err := mongodb.AddTeachersToDb(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: addedTeachers}, nil
}

func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeachersRequest) (*pb.Teachers, error) {
	// Getting all the filters
	filters, err := buildFilterForModel(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Getting all the sorting options
	sortOptions := buildSortOptions(req.SortBy)

	// Querying the database
	teachers, err := mongodb.GetTeachersFromDB(ctx, sortOptions, filters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: teachers}, nil

}

func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	updatedTeachers, err := mongodb.ModifyTeachersInDB(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: updatedTeachers}, nil
}

func (s *Server) DeleteTeachers(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeachersConfirmation, error) {
	ids := req.GetIds()
	var objectIdsToDelete []primitive.ObjectID
	for _, v := range ids {
		if v.Id == "" {
			return nil, status.Error(codes.InvalidArgument, errors.New("ID field cannot be empty").Error())
		}
		objId, err := primitive.ObjectIDFromHex(v.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		objectIdsToDelete = append(objectIdsToDelete, objId)
	}

	deletedIds, err := mongodb.DeleteTeachersFromDB(ctx, objectIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteTeachersConfirmation{
		Status:     "Teachers successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}
