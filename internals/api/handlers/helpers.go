package handlers

import (
	"ClassConnectRPC/pkg/utils"
	"reflect"
	"strings"

	pb "ClassConnectRPC/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func buildFilterForModel(object interface{}, model interface{}) (bson.M, error) {
	filter := bson.M{}

	modelVal := reflect.ValueOf(model).Elem()
	modelType := modelVal.Type()

	reqVal := reflect.ValueOf(object).Elem()
	reqType := reqVal.Type()

	for i := 0; i < reqVal.NumField(); i++ {
		fieldVal := reqVal.Field(i)
		fieldName := reqType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			modelField := modelVal.FieldByName(fieldName)
			if modelField.IsValid() && modelField.CanSet() {
				modelField.Set(fieldVal)
			}
		}
	}

	// We iterate over modelTeacher to build our filter
	for i := 0; i < modelVal.NumField(); i++ {
		fieldVal := modelVal.Field(i)
		fieldName := modelType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			bsonTag := modelType.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")

			// We cannot filter by '_id' as it is a plain string and mongodb stores an 'ObjectID' instance as the '_id' field
			if bsonTag == "_id" {
				// Accepts a string and returns a mongo object id
				objId, err := primitive.ObjectIDFromHex(reqVal.FieldByName(fieldName).Interface().(string))
				if err != nil {
					return nil, utils.ErrorHandler(err, "Invalid ID")
				}
				filter[bsonTag] = objId
			} else {
				filter[bsonTag] = fieldVal.Interface().(string)
			}

		}
	}

	return filter, nil
}

func buildSortOptions(sortFields []*pb.SortField) bson.D {
	var sortOptions bson.D

	for _, sortField := range sortFields {
		// 1 represents ascending order (default is ascending)
		order := 1
		if sortField.GetOrder() == pb.Order_DSC {
			order = -1
		}
		sortOptions = append(sortOptions, bson.E{Key: sortField.Field, Value: order})
	}

	return sortOptions
}
