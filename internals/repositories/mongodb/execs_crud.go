package mongodb

import (
	"ClassConnectRPC/internals/models"
	"ClassConnectRPC/pkg/utils"
	pb "ClassConnectRPC/proto/gen"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddExecsToDb(ctx context.Context, execsFromRequest []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to mongodb")
	}
	defer client.Disconnect(ctx)

	newExecs := make([]*models.Exec, len(execsFromRequest))
	for i, pbExec := range execsFromRequest {
		modelExec := MapPbExecToModelExec(pbExec)

		// Hash the password before storing
		if modelExec.Password != "" {
			hashedPassword, err := utils.HashPassword(modelExec.Password)
			if err != nil {
				return nil, utils.ErrorHandler(err, "Error hashing password")
			}
			modelExec.Password = hashedPassword
		}

		newExecs[i] = modelExec
	}

	var addedExecs []*pb.Exec
	for _, exec := range newExecs {
		res, err := client.Database("school").Collection("execs").InsertOne(ctx, exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into mongodb")
		}

		objectId, ok := res.InsertedID.(primitive.ObjectID)
		if ok {
			exec.Id = objectId.Hex()
		}

		pbExec := MapModelExecToPbExec(exec)

		addedExecs = append(addedExecs, pbExec)
	}
	return addedExecs, nil
}

func GetExecsFromDB(ctx context.Context, sortOptions bson.D, filters bson.M) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	coll := client.Database("school").Collection("execs")
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

	execs, err := decodeEntities(ctx, cursor, func() *pb.Exec { return &pb.Exec{} }, func() *models.Exec { return &models.Exec{} })
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	return execs, nil
}

func ModifyExecsInDB(ctx context.Context, req *pb.Execs) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	var updatedExecs []*pb.Exec
	for _, exec := range req.Execs {
		if exec.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id cannot be blank"), "id cannot be blank")
		}

		modelExec := MapPbExecToModelExec(exec)

		// Hash the password if it's being updated
		if modelExec.Password != "" {
			hashedPassword, err := utils.HashPassword(modelExec.Password)
			if err != nil {
				return nil, utils.ErrorHandler(err, "Error hashing password")
			}
			modelExec.Password = hashedPassword
		}

		objId, err := primitive.ObjectIDFromHex(exec.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid ID")
		}

		// Convert modelExec into a BSON document before updating the database
		modelDoc, err := bson.Marshal(modelExec)
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

		_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("Error updating exec with ID: %s", exec.Id))
		}

		updatedExec := MapModelExecToPbExec(modelExec)
		updatedExecs = append(updatedExecs, updatedExec)
	}
	return updatedExecs, nil
}

func DeleteExecsFromDB(ctx context.Context, objectIdsToDelete []primitive.ObjectID) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"_id": bson.M{"$in": objectIdsToDelete}}
	result, err := client.Database("school").Collection("execs").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "No execs were deleted")
	}

	var deletedIds []string
	for _, v := range objectIdsToDelete {
		deletedIds = append(deletedIds, v.Hex())
	}
	return deletedIds, nil
}

func GetUserByUsername(ctx context.Context, username string) (*models.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	filter := bson.M{"username": username}

	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, filter).Decode(&exec)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "User not found. Incorrect username or password ")
		}
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	return &exec, nil
}

func UpdateUserInDB(ctx context.Context, req *pb.UpdatePasswordRequest) (*models.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	var exec models.Exec
	err = client.Database("school").Collection("execs").FindOne(ctx, bson.M{"_id": objId}).Decode(&exec)
	if err != nil {
		return nil, utils.ErrorHandler(err, "User not found")
	}

	err = utils.VerifyPassword(req.GetCurrentPassword(), exec.Password)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Invalid username or password")
	}

	newHashedPassword, err := utils.HashPassword(req.GetNewPassword())
	if err != nil {
		return nil, utils.ErrorHandler(err, "Unable to hash the password")
	}

	update := bson.M{
		"$set": bson.M{
			"password":            newHashedPassword,
			"password_changed_at": time.Now().Format(time.RFC3339),
		},
	}
	_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"_id": objId}, update)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to update the password")
	}
	return &exec, nil
}

func DeactivateUserInDB(ctx context.Context, objIds []primitive.ObjectID) (*pb.Confirmation, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")
	}
	defer client.Disconnect(ctx)

	update := bson.M{"$set": bson.M{"inactive_status": true}}
	_, err = client.Database("school").Collection("execs").UpdateMany(ctx, bson.M{"_id": bson.M{"$in": objIds}}, update)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	return nil, nil
}
