package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naki/location-service/functions/api_functions"
	"github.com/naki/location-service/models"
)

func UpdateLocation(c *gin.Context) {
	nurseID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "invalid user id"})
		return
	}

	var body models.LocationUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "latitude and longitude required"})
		return
	}

	if err := api_functions.ProcessLocationUpdate(nurseID, body); err != nil {
		log.Printf("location update failed for nurse %s: %v", nurseID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "failed to update location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "location updated",
	})
}

func GetNurseLocation(c *gin.Context) {
	nurseID, err := uuid.Parse(c.Param("nurse_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "invalid nurse id"})
		return
	}

	loc, err := api_functions.GetNurseLocation(nurseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": "nurse location not available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "nurse location retrieved",
		"data":    loc,
	})
}

func TrackVisit(c *gin.Context) {
	visitID := c.Param("visit_id")

	info, err := api_functions.GetTrackingInfo(visitID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "tracking info retrieved",
		"data":    info,
	})
}

func GetActiveNurses(c *gin.Context) {
	locations, err := api_functions.GetAllNurseLocations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error()})
		return
	}

	if locations == nil {
		locations = []models.NurseLocation{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "active nurse locations retrieved",
		"data":    locations,
		"count":   len(locations),
	})
}

func GetActiveVisits(c *gin.Context) {
	visits, err := api_functions.GetAllActiveVisits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error()})
		return
	}

	if visits == nil {
		visits = []models.ActiveVisit{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "active visits retrieved",
		"data":    visits,
		"count":   len(visits),
	})
}

func GetLocationHistory(c *gin.Context) {
	nurseID, err := uuid.Parse(c.Param("nurse_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "invalid nurse id"})
		return
	}

	logs, err := api_functions.GetLocationHistory(nurseID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "location history retrieved",
		"data":    logs,
	})
}

func GetVisitTrail(c *gin.Context) {
	visitID := c.Param("visit_id")

	logs, err := api_functions.GetVisitTrail(visitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": err.Error()})
		return
	}

	if logs == nil {
		logs = []models.LocationLog{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "visit trail retrieved",
		"data":    logs,
		"count":   len(logs),
	})
}
