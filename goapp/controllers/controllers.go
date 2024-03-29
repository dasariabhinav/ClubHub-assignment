package controllers

import (
	"context"
	"fmt"
	"net/http"
	"project/database"
	"project/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetData(c *gin.Context) {
	var items []models.Data

	cursor, err := database.Collection.Find(context.Background(), bson.D{})
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var item models.Data
		if err := cursor.Decode(&item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode items"})
			return
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, items)

}

func CreateData(c *gin.Context) {
	var newItem models.Data

	// Bind the JSON data from the request body to the newData variable
	if err := c.ShouldBindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Iterate through franchises and fetch WHOIS data
	for i := range newItem.Company.Franchises {
		franchise := &newItem.Company.Franchises[i]
		whoisData, err := getWHOISData(franchise.URL)
		if err != nil {
			fmt.Println("Error fetching WHOIS data:", err)
			// Handle the error as needed (e.g., log it, but continue)
		} else {
			franchise.WhoIsInfo = whoisData
		}
	}

	// Insert into MongoDB
	insertResult, err := database.Collection.InsertOne(context.Background(), newItem)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"insertedID": insertResult.InsertedID})
}

func UpdateData(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	var updateData models.Data

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the document with the specified ID
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updateData}

	result, err := database.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data"})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No document found for the given ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

func GetDataByFranchiseName(c *gin.Context) {
	// Get the franchise name from the request parameters
	franchiseName := c.Param("franchise_name")

	// Define a filter to search for documents with the specified franchise name
	filter := bson.M{"company.franchises.name": franchiseName}

	// Find documents matching the filter
	var items []models.Data
	cursor, err := database.Collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}
	defer cursor.Close(context.Background())

	// Decode the documents into a slice of Data structs
	for cursor.Next(context.Background()) {
		var item models.Data
		if err := cursor.Decode(&item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode items"})
			return
		}
		items = append(items, item)
	}

	// Check if any items were found
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No documents found for the given franchise name"})
		return
	}

	c.JSON(http.StatusOK, items)
}
