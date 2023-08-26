package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
)

func setupDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	database = client.Database("students")
	collection = database.Collection("students")
}

type Student struct {
	Firstname  string `json:"firstname" bson:"firstname"`
	Lastname   string `json:"lastname" bson:"lastname"`
	Age        int64  `json:"age" bson:"age"`
	Department string `json:"department" bson:"department"`
	EmailId    string `json:"emailId" bson:"emailId"`
}

func registerStudent(c *gin.Context) {
	var student Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.InsertOne(ctx, student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func getStudent(c *gin.Context) {
	emailID := c.Param("emailId") // Retrieve the emailId parameter from the path
	var student Student
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, bson.M{"emailId": emailID}).Decode(&student)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, student)
}

func updateStudentDetails(c *gin.Context) {
	emailID := c.Param("emailId")
	var updatedStudent Student
	if err := c.ShouldBindJSON(&updatedStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.M{
		"$set": bson.M{
			"firstname":  updatedStudent.Firstname,
			"lastname":   updatedStudent.Lastname,
			"age":        updatedStudent.Age,
			"department": updatedStudent.Department,
			"emailId":    updatedStudent.EmailId,
		},
	}
	result, err := collection.UpdateOne(ctx, bson.M{"emailId": emailID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func deleteStudentHandler(c *gin.Context) {
	emailID := c.Param("emailId")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.DeleteOne(ctx, bson.M{"emailId": emailID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func getAllStudents(c *gin.Context) {
	var students []Student
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &students); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, students)
}

func main() {
	router := gin.Default()
	setupDB()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"data": "Hello from NexaScale",
		})
	})

	router.POST("/register", registerStudent)
	router.GET("/student/:emailId", getStudent)
	router.GET("/students", getAllStudents)
	router.PUT("/update/:emailId", updateStudentDetails)
	router.DELETE("delete/:emailId", deleteStudentHandler)

	router.Run("localhost:6000")

}
