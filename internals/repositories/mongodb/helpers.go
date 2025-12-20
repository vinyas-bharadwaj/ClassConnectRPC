package mongodb

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/pkg/utils"
	pb "ClassConnectRPC/proto/gen"
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
)

// This is a generic function which works for teachers, students and execs
// Same can be implemented using interfaces instead of generics
// General rule of thumb is to use generics when the return type is dependent on the function arguments
// If the return type is independent of the function arguments, it's better to use interfaces
func decodeEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newEntity func() *T, newModel func() *M) ([]*T, error) {
	var entities []*T
	for cursor.Next(ctx) {
		model := newModel()
		err := cursor.Decode(&model)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		entity := newEntity()
		modelVal := reflect.ValueOf(model).Elem()
		pbVal := reflect.ValueOf(entity).Elem()

		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldName := modelVal.Type().Field(i).Name

			pbField := pbVal.FieldByName(modelFieldName)
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}

		entities = append(entities, entity)
	}
	return entities, nil
}

func MapModelToPb[M any, P any](model *M, newPb func() *P) *P {
	pbEntity := newPb()
	modelVal := reflect.ValueOf(model).Elem()
	pbVal := reflect.ValueOf(pbEntity).Elem()

	for i := 0; i < modelVal.NumField(); i++ {
		modelField := modelVal.Field(i)
		modelFieldType := modelVal.Type().Field(i)

		pbField := pbVal.FieldByName(modelFieldType.Name)
		if pbField.IsValid() && pbField.CanSet() {
			pbField.Set(modelField)
		}

	}
	return pbEntity
}

func MapModelTeacherToPbTeacher(teacherModel *models.Teacher) *pb.Teacher {
	return MapModelToPb(teacherModel, func() *pb.Teacher { return &pb.Teacher{} })
}

func MapPbToModel[P any, M any](pbStruct *P, newModel func() *M) *M {
	modelEntity := newModel()
	pbVal := reflect.ValueOf(pbStruct).Elem()
	modelVal := reflect.ValueOf(modelEntity).Elem()

	for j := 0; j < pbVal.NumField(); j++ {
		pbField := pbVal.Field(j)
		fieldName := pbVal.Type().Field(j).Name

		modelField := modelVal.FieldByName(fieldName)
		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		}
	}

	return modelEntity
}

func MapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	return MapPbToModel(pbTeacher, func() *models.Teacher { return &models.Teacher{} })
}
