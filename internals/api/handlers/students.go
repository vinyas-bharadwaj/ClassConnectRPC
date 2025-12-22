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

func (s *Server) AddStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {
	for _, student := range req.GetStudents() {
		if student.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "Request is in incorrect format: non-empty ID field is not allowed")
		}
	}

	addedStudents, err := mongodb.AddStudentsToDb(ctx, req.GetStudents())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: addedStudents}, nil
}

func (s *Server) GetStudents(ctx context.Context, req *pb.GetStudentsRequest) (*pb.Students, error) {
	// Getting all the filters
	filters, err := buildFilterForModel(req.Student, &models.Student{})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Getting all the sorting options
	sortOptions := buildSortOptions(req.SortBy)

	// Querying the database
	students, err := mongodb.GetStudentsFromDB(ctx, sortOptions, filters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: students}, nil

}

func (s *Server) UpdateStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {
	updatedStudents, err := mongodb.ModifyStudentsInDB(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Students{Students: updatedStudents}, nil
}

func (s *Server) DeleteStudents(ctx context.Context, req *pb.StudentIds) (*pb.DeleteStudentsConfirmation, error) {
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

	deletedIds, err := mongodb.DeleteStudentsFromDB(ctx, objectIdsToDelete)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteStudentsConfirmation{
		Status:     "Students successfully deleted",
		DeletedIds: deletedIds,
	}, nil
}
