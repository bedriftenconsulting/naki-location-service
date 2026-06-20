package api_functions

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/naki/location-service/conf"
	"github.com/naki/location-service/models"
	"github.com/segmentio/kafka-go"
)

type NurseMatchedEvent struct {
	BookingID     string  `json:"booking_id"`
	CustomerID    string  `json:"customer_id"`
	NurseID       string  `json:"nurse_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	ServiceType   string  `json:"service_type"`
	ScheduledAt   string  `json:"scheduled_at"`
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Distance      float64 `json:"distance_km"`
}

type VisitCompletedEvent struct {
	BookingID string `json:"booking_id"`
	VisitID   string `json:"visit_id"`
	NurseID   string `json:"nurse_id"`
}

func StartKafkaConsumers() {
	go consumeNurseMatched()
	go consumeVisitCompleted()
}

func consumeNurseMatched() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{conf.AppConfig.KafkaBroker},
		Topic:    "nurse.matched",
		GroupID:  "location-service",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("listening on kafka topic: nurse.matched")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading from topic nurse.matched: %v", err)
			continue
		}

		var event NurseMatchedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal nurse.matched event: %v", err)
			continue
		}

		log.Printf("nurse.matched received: booking=%s nurse=%s", event.BookingID, event.NurseID)

		visit := models.ActiveVisit{
			VisitID:      event.BookingID,
			BookingID:    event.BookingID,
			NurseID:      event.NurseID,
			CustomerID:   event.CustomerID,
			PatientLat:   event.Latitude,
			PatientLng:   event.Longitude,
			PatientName:  event.CustomerName,
			PatientPhone: event.CustomerPhone,
			ServiceType:  event.ServiceType,
			StartedAt:    time.Now(),
		}

		if err := SetActiveVisit(visit); err != nil {
			log.Printf("failed to set active visit: %v", err)
			continue
		}

		log.Printf("active visit created: %s (nurse=%s → patient at %.4f,%.4f)",
			visit.VisitID, visit.NurseID, visit.PatientLat, visit.PatientLng)
	}
}

func consumeVisitCompleted() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{conf.AppConfig.KafkaBroker},
		Topic:    "visit.completed",
		GroupID:  "location-service",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("listening on kafka topic: visit.completed")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("error reading from topic visit.completed: %v", err)
			continue
		}

		var event VisitCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal visit.completed event: %v", err)
			continue
		}

		visitID := event.VisitID
		if visitID == "" {
			visitID = event.BookingID
		}

		log.Printf("visit.completed received: visit=%s nurse=%s", visitID, event.NurseID)

		CloseVisitTracking(visitID)

		if err := RemoveActiveVisit(visitID, event.NurseID); err != nil {
			log.Printf("failed to remove active visit: %v", err)
		}

		log.Printf("visit tracking stopped: %s", visitID)
	}
}
