package api_functions

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/naki/location-service/database"
	"github.com/naki/location-service/models"
)

const earthRadiusKm = 6371.0

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func ProcessLocationUpdate(nurseID uuid.UUID, update models.LocationUpdate) error {
	loc := models.NurseLocation{
		NurseID:   nurseID,
		Latitude:  update.Latitude,
		Longitude: update.Longitude,
		Heading:   update.Heading,
		Speed:     update.Speed,
		VisitID:   update.VisitID,
		UpdatedAt: time.Now(),
	}

	if err := SetNurseLocation(loc); err != nil {
		return fmt.Errorf("failed to cache location: %w", err)
	}

	go logLocationToDB(nurseID, update)

	if update.VisitID != "" {
		go notifyTrackingClients(update.VisitID, loc)
	}

	return nil
}

func logLocationToDB(nurseID uuid.UUID, update models.LocationUpdate) {
	query := `
		INSERT INTO location_logs (nurse_id, visit_id, latitude, longitude, heading, speed)
		VALUES ($1, $2, $3, $4, $5, $6)`

	var visitID *string
	if update.VisitID != "" {
		visitID = &update.VisitID
	}

	_, err := database.DB.Exec(query, nurseID, visitID,
		update.Latitude, update.Longitude, update.Heading, update.Speed)
	if err != nil {
		log.Printf("failed to log location: %v", err)
	}
}

func GetTrackingInfo(visitID string) (*models.TrackingResponse, error) {
	visit, err := GetActiveVisit(visitID)
	if err != nil {
		return nil, fmt.Errorf("visit not found or not active")
	}

	nurseID, err := uuid.Parse(visit.NurseID)
	if err != nil {
		return nil, fmt.Errorf("invalid nurse id")
	}

	nurseLoc, err := GetNurseLocation(nurseID)
	if err != nil {
		return nil, fmt.Errorf("nurse location not available")
	}

	dist := haversine(nurseLoc.Latitude, nurseLoc.Longitude, visit.PatientLat, visit.PatientLng)
	dist = math.Round(dist*100) / 100

	eta := estimateETA(dist, nurseLoc.Speed)

	return &models.TrackingResponse{
		NurseLocation: nurseLoc,
		Visit:         visit,
		ETA:           eta,
		DistanceKm:    dist,
	}, nil
}

func estimateETA(distanceKm float64, speedKmh float64) string {
	if speedKmh < 5 {
		speedKmh = 30
	}

	minutes := (distanceKm / speedKmh) * 60

	if minutes < 1 {
		return "Less than 1 minute"
	} else if minutes < 60 {
		return fmt.Sprintf("%.0f minutes", math.Ceil(minutes))
	} else {
		hours := int(minutes / 60)
		mins := int(minutes) % 60
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
}

func GetLocationHistory(nurseID uuid.UUID, limit int) ([]models.LocationLog, error) {
	var logs []models.LocationLog
	query := `SELECT * FROM location_logs WHERE nurse_id = $1 ORDER BY created_at DESC LIMIT $2`
	err := database.DB.Select(&logs, query, nurseID, limit)
	return logs, err
}

func GetVisitTrail(visitID string) ([]models.LocationLog, error) {
	var logs []models.LocationLog
	query := `SELECT * FROM location_logs WHERE visit_id = $1 ORDER BY created_at ASC`
	err := database.DB.Select(&logs, query, visitID)
	return logs, err
}
