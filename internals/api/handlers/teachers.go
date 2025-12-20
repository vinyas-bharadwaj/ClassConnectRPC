package handlers

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/internals/repositories/mongodb"
	pb "ClassConnectRPC/proto/gen"
	"context"

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
