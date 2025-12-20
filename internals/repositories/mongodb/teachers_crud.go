package mongodb

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/pkg/utils"
	"context"
	"errors"
	"fmt"

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
		modelTeacher := MapPbTeacherToModelTeacher(pbTeacher)
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

		pbTeacher := MapModelTeacherToPbTeacher(teacher)

		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
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

	teachers, err := decodeEntities(ctx, cursor, func() *pb.Teacher { return &pb.Teacher{} }, func() *models.Teacher { return &models.Teacher{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	return teachers, nil
}

func ModifyTeachersInDB(ctx context.Context, req *pb.Teachers) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher
	for _, teacher := range req.Teachers {
		if teacher.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id cannot be blank"), "id cannot be blank")
		}

		modelTeacher := MapPbTeacherToModelTeacher(teacher)
		objId, err := primitive.ObjectIDFromHex(teacher.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}

		// Convert modelTeacher into a BSON document before updating the database
		modelDoc, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal error")
		}

		// Remove the '_id' field from 'updateDoc' since we are not meant to update the id
		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("teachers").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("Error updating teacher with ID: %s", teacher.Id))
		}

		updatedTeacher := MapModelTeacherToPbTeacher(modelTeacher)
		updatedTeachers = append(updatedTeachers, updatedTeacher)
	}
	return updatedTeachers, nil
}

func DeleteTeachersFromDB(ctx context.Context, objectIdsToDelete []primitive.ObjectID) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"_id": bson.M{"$in": objectIdsToDelete}}
	result, err := client.Database("school").Collection("teachers").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "No teachers were deleted")
	}

	var deletedIds []string
	for _, v := range objectIdsToDelete {
		deletedIds = append(deletedIds, v.Hex())
	}
	return deletedIds, nil
}
