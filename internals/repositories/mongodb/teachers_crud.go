package mongodb

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/pkg/utils"
	"context"
	"reflect"

	pb "ClassConnectRPC/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddTeachersToDb(ctx context.Context, teachersFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to mongodb")
	}
	defer client.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, len(teachersFromReq))
	for i, pbTeacher := range teachersFromReq {
		modelTeacher := mapPbTeacherToModelTeacher(pbTeacher)
		newTeachers[i] = modelTeacher
	}

	var addedTeachers []*pb.Teacher
	for _, teacher := range newTeachers {
		res, err := client.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into mongodb")
		}

		objectId, ok := res.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectId.Hex()
		}

		pbTeacher := mapModelTeacherToPbTeacher(teacher)

		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
}

func mapModelTeacherToPbTeacher(teacher *models.Teacher) *pb.Teacher {
	pbTeacher := &pb.Teacher{}
	modelVal := reflect.ValueOf(teacher).Elem()
	pbVal := reflect.ValueOf(pbTeacher).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		modelFieldType := modelVal.Type().Field(i)

		pbField := pbVal.FieldByName(modelFieldType.Name)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		}

	}
	return pbTeacher
}

func mapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	modelTeacher := models.Teacher{}
	pbVal := reflect.ValueOf(pbTeacher).Elem()
	modelVal := reflect.ValueOf(&modelTeacher).Elem()

	for j := 0; j < pbVal.NumField(); j++ {
		pbField := pbVal.Field(j)
		fieldName := pbVal.Type().Field(j).Name

		modelField := modelVal.FieldByName(fieldName)
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		}
	}

	return &modelTeacher
}

func GetTeachersFromDB(ctx context.Context, sortOptions bson.D, filters bson.M) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("teachers")
	var cursor *mongo.Cursor
	if len(sortOptions) >= 1 {
		cursor, err = coll.Find(ctx, filters, options.Find().SetSort(sortOptions))
	} else {
		cursor, err = coll.Find(ctx, filters)
	}
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer cursor.Close(ctx)

	var teachers []*pb.Teacher
	for cursor.Next(ctx) {
		var teacher models.Teacher
		err = cursor.Decode(&teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		teachers = append(teachers, &pb.Teacher{
			Id:        teacher.Id,
			FirstName: teacher.FirstName,
			LastName:  teacher.LastName,
			Email:     teacher.Email,
			Class:     teacher.Class,
			Subject:   teacher.Subject,
		})
	}
	return teachers, nil
}
