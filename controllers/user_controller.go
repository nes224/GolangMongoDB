package controllers

import (
	"context"
	"fmt"
	_ "fmt"
	"gin-mongo-api/configs"
	"gin-mongo-api/models"
	"gin-mongo-api/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")
var validate = validator.New()

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var user models.User
		defer cancel()
		// validate the request body
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}
		fmt.Printf("Result: %+v\n",user)
		//use the validator library to validate required fields
		if validationErr := validate.Struct(&user); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}
		newUser := models.User{
			Id:       primitive.NewObjectID(),
			Name:     user.Name,
			Location: user.Location,
			Title:    user.Title,
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
		}

		response := map[string]interface{}{
			"_id":      result.InsertedID,
			"name":     user.Name,
			"location": user.Location,
			"title":    user.Title,
		}
		c.JSON(http.StatusCreated, 
			responses.UserResponse{
				Status: http.StatusCreated, 
				Message: "success", 
				Data: response,
			})
	}
}

func GetAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		var user models.User
		defer cancel()
		objId, _ := primitive.ObjectIDFromHex(userId)
		err := userCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"Data": err.Error()}})
			return
		}
		result := &user
		c.JSON(http.StatusOK, 
			responses.UserResponse{
				Status: http.StatusOK, 
				Message: "success", 
				Data: map[string]interface{}{"Data": result},
		})
	}
}

func EditAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		var user models.User
		defer cancel()
		objId, _ := primitive.ObjectIDFromHex(userId)

		//validate the request body
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&user); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}

		update := bson.M{"name": user.Name, "localtion": user.Location, "title": user.Title}
		result, err := userCollection.UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"Data": err.Error()}})
			return
		}
		//get updated user details
		var updatedUser models.User
		if result.MatchedCount == 1 {
			err := userCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&updatedUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"Data": err.Error()}})
				return
			}
		}
		c.JSON(http.StatusOK, 
			responses.UserResponse{
				Status: http.StatusInternalServerError, 
				Message: "Success", 
				Data: map[string]interface{}{"Data": updatedUser,
			},
		})
	}
}

func DeleteAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(userId)

		result, err := userCollection.DeleteOne(ctx, bson.M{"_id": objId})
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"Data": err.Error()}})
			return
		}

		if result.DeletedCount < 1 {
			c.JSON(http.StatusNotFound,
				responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": "User with specified ID not found!"}},
			)
			return
		}
		c.JSON(http.StatusOK,
			responses.UserResponse{
				Status:  http.StatusOK,
				Message: "Success",
				Data:    map[string]interface{}{"Data": "User successfully deleted!"}},
		)
	}
}

func GetAllUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var users []models.User
		defer cancel()

		results, err := userCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error", Data: map[string]interface{}{"Data": err.Error()}})
			return
		}

		//reading from the db in an optional way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleUser models.User
			if err = results.Decode(&singleUser); err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"Data": err.Error()}})
				return
			}
			users = append(users, singleUser)
		}
		c.JSON(http.StatusOK,
			responses.UserResponse{Status: http.StatusOK, 
				Message: "success", 
				Data: map[string]interface{}{"Data": users}},
		)
	}
}

type searchUsers struct {
	offsetSize int32 `form:"offset" binding:"required,min=1"`
	limit int32 `form:"limit" binding:"required,min=5,max=10"`
}

func SearchUser() gin.HandlerFunc{
	return func(c *gin.Context){
		var req searchUsers
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		
		var users[] models.User
		defer cancel()
		l := int64(req.limit)
		skip := int64(req.offsetSize)
		fmt.Println("L: l", req.limit)
		fmt.Println("Offset: skip", req.offsetSize)
		fOpt := options.FindOptions{Limit: &l, Skip: &skip}
		results, err := userCollection.Find(ctx, bson.M{}, &fOpt)
		if err != nil {
			c.JSON(http.StatusInternalServerError,responses.UserResponse{Status: http.StatusInternalServerError,Message:"Error"})
			return 
		}
		// fmt.Println("results %+v\n", c.BindJSON(results))
		defer results.Close(ctx)
		for results.Next(ctx) {
			var result models.User
			if err := results.Decode(&result); err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"Data": err.Error()}})
				return 
			}
			users = append(users, result)
		}
		c.JSON(http.StatusOK,
			responses.UserResponse{Status: http.StatusOK, 
				Message: "success", 
				Data: map[string]interface{}{"Data": users}},
		)
	}
}
