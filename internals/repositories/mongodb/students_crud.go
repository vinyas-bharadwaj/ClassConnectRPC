package mongodb

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/pkg/utils"
	pb "ClassConnectRPC/proto/gen"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddStudentsToDb(ctx context.Context, studentsFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to mongodb")
	}
	defer client.Disconnect(ctx)

	newStudents := make([]*models.Student, len(studentsFromReq))
	for i, pbStudent := range studentsFromReq {
		modelStudent := MapPbStudentToModelStudent(pbStudent)
		newStudents[i] = modelStudent
	}

	var addedStudents []*pb.Student
	for _, student := range newStudents {
		res, err := client.Database("school").Collection("students").InsertOne(ctx, student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into mongodb")
		}

		objectId, ok := res.InsertedID.(primitive.ObjectID)
		if ok {
			student.Id = objectId.Hex()
		}

		pbStudent := MapModelStudentToPbStudent(student)

		addedStudents = append(addedStudents, pbStudent)
	}
	return addedStudents, nil
}

func GetStudentsFromDB(ctx context.Context, sortOptions bson.D, filters bson.M) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("students")
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

	students, err := decodeEntities(ctx, cursor, func() *pb.Student { return &pb.Student{} }, func() *models.Student { return &models.Student{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	return students, nil
}

func ModifyStudentsInDB(ctx context.Context, req *pb.Students) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	var updatedStudents []*pb.Student
	for _, student := range req.Students {
		if student.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id cannot be blank"), "id cannot be blank")
		}

		modelStudent := MapPbStudentToModelStudent(student)
		objId, err := primitive.ObjectIDFromHex(student.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}

		// Convert modelStudent into a BSON document before updating the database
		modelDoc, err := bson.Marshal(modelStudent)
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

		_, err = client.Database("school").Collection("students").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("Error updating student with ID: %s", student.Id))
		}

		updatedStudent := MapModelStudentToPbStudent(modelStudent)
		updatedStudents = append(updatedStudents, updatedStudent)
	}
	return updatedStudents, nil
}

func DeleteStudentsFromDB(ctx context.Context, objectIdsToDelete []primitive.ObjectID) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"_id": bson.M{"$in": objectIdsToDelete}}
	result, err := client.Database("school").Collection("students").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "No students were deleted")
	}

	var deletedIds []string
	for _, v := range objectIdsToDelete {
		deletedIds = append(deletedIds, v.Hex())
	}
	return deletedIds, nil
}

func GetStudentsByTeacherIdFromDB(ctx context.Context, teacherId string) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Invalid ID")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objId}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "Teacher with given ID not found")
		}
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	cursor, err := client.Database("school").Collection("students").Find(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer cursor.Close(ctx)

	students, err := decodeEntities(ctx, cursor, func() *pb.Student { return &pb.Student{} }, func() *models.Student { return &models.Student{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	return students, nil
}

func GetStudentCountByTeacherIdFromDB(ctx context.Context, teacherId string) (int64, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return 0, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Invalid ID")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objId}).Decode(&teacher)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Internal error")
	}

	count, err := client.Database("school").Collection("students").CountDocuments(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return 0, utils.ErrorHandler(err, "Internal error")
	}
	return count, nil
}
