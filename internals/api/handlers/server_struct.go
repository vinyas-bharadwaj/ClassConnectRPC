package handlers

import pb "ClassConnectRPC/proto/gen"

type Server struct {
	pb.UnimplementedTeachersServiceServer
	pb.UnimplementedStudentsSerciesServer
	pb.UnimplementedExecsServiceServer
}
