package models

import (
	"time"

	"github.com/google/uuid"
)

type NurseLocation struct {
	NurseID   uuid.UUID `json:"nurse_id"   redis:"nurse_id"`
	Latitude  float64   `json:"latitude"   redis:"latitude"`
	Longitude float64   `json:"longitude"  redis:"longitude"`
	Heading   float64   `json:"heading"    redis:"heading"`
	Speed     float64   `json:"speed"      redis:"speed"`
	VisitID   string    `json:"visit_id"   redis:"visit_id"`
	UpdatedAt time.Time `json:"updated_at" redis:"updated_at"`
}

type ActiveVisit struct {
	VisitID      string    `json:"visit_id"       redis:"visit_id"`
	BookingID    string    `json:"booking_id"     redis:"booking_id"`
	NurseID      string    `json:"nurse_id"       redis:"nurse_id"`
	CustomerID   string    `json:"customer_id"    redis:"customer_id"`
	PatientLat   float64   `json:"patient_lat"    redis:"patient_lat"`
	PatientLng   float64   `json:"patient_lng"    redis:"patient_lng"`
	PatientName  string    `json:"patient_name"   redis:"patient_name"`
	PatientPhone string    `json:"patient_phone"  redis:"patient_phone"`
	ServiceType  string    `json:"service_type"   redis:"service_type"`
	StartedAt    time.Time `json:"started_at"     redis:"started_at"`
}

type LocationLog struct {
	ID        uuid.UUID `db:"id"          json:"id"`
	NurseID   uuid.UUID `db:"nurse_id"    json:"nurse_id"`
	VisitID   *string   `db:"visit_id"    json:"visit_id"`
	Latitude  float64   `db:"latitude"    json:"latitude"`
	Longitude float64   `db:"longitude"   json:"longitude"`
	Heading   float64   `db:"heading"     json:"heading"`
	Speed     float64   `db:"speed"       json:"speed"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
}

type LocationUpdate struct {
	Latitude  float64 `json:"latitude"  binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
	VisitID   string  `json:"visit_id"`
}

type TrackingResponse struct {
	NurseLocation *NurseLocation `json:"nurse_location"`
	Visit         *ActiveVisit   `json:"visit"`
	ETA           string         `json:"eta"`
	DistanceKm    float64        `json:"distance_km"`
}
