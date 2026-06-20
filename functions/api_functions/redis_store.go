package api_functions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/naki/location-service/conf"
	"github.com/naki/location-service/models"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     conf.AppConfig.RedisAddr,
		Password: conf.AppConfig.RedisPass,
		DB:       1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: redis not available: %v", err)
	} else {
		log.Println("redis connected successfully")
	}
}

func nurseLocationKey(nurseID uuid.UUID) string {
	return fmt.Sprintf("location:nurse:%s", nurseID.String())
}

func activeVisitKey(visitID string) string {
	return fmt.Sprintf("location:visit:%s", visitID)
}

func nurseVisitKey(nurseID string) string {
	return fmt.Sprintf("location:nurse_visit:%s", nurseID)
}

func SetNurseLocation(loc models.NurseLocation) error {
	ctx := context.Background()

	data, err := json.Marshal(loc)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, nurseLocationKey(loc.NurseID), data, 5*time.Minute).Err()
}

func GetNurseLocation(nurseID uuid.UUID) (*models.NurseLocation, error) {
	ctx := context.Background()

	data, err := rdb.Get(ctx, nurseLocationKey(nurseID)).Bytes()
	if err != nil {
		return nil, err
	}

	var loc models.NurseLocation
	if err := json.Unmarshal(data, &loc); err != nil {
		return nil, err
	}

	return &loc, nil
}

func GetAllNurseLocations() ([]models.NurseLocation, error) {
	ctx := context.Background()

	keys, err := rdb.Keys(ctx, "location:nurse:*").Result()
	if err != nil {
		return nil, err
	}

	var locations []models.NurseLocation

	for _, key := range keys {
		data, err := rdb.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var loc models.NurseLocation
		if err := json.Unmarshal(data, &loc); err != nil {
			continue
		}

		locations = append(locations, loc)
	}

	return locations, nil
}

func RemoveNurseLocation(nurseID uuid.UUID) error {
	ctx := context.Background()
	return rdb.Del(ctx, nurseLocationKey(nurseID)).Err()
}

func SetActiveVisit(visit models.ActiveVisit) error {
	ctx := context.Background()

	data, err := json.Marshal(visit)
	if err != nil {
		return err
	}

	pipe := rdb.Pipeline()
	pipe.Set(ctx, activeVisitKey(visit.VisitID), data, 4*time.Hour)
	pipe.Set(ctx, nurseVisitKey(visit.NurseID), visit.VisitID, 4*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func GetActiveVisit(visitID string) (*models.ActiveVisit, error) {
	ctx := context.Background()

	data, err := rdb.Get(ctx, activeVisitKey(visitID)).Bytes()
	if err != nil {
		return nil, err
	}

	var visit models.ActiveVisit
	if err := json.Unmarshal(data, &visit); err != nil {
		return nil, err
	}

	return &visit, nil
}

func GetActiveVisitByNurse(nurseID string) (*models.ActiveVisit, error) {
	ctx := context.Background()

	visitID, err := rdb.Get(ctx, nurseVisitKey(nurseID)).Result()
	if err != nil {
		return nil, err
	}

	return GetActiveVisit(visitID)
}

func RemoveActiveVisit(visitID string, nurseID string) error {
	ctx := context.Background()

	pipe := rdb.Pipeline()
	pipe.Del(ctx, activeVisitKey(visitID))
	pipe.Del(ctx, nurseVisitKey(nurseID))
	_, err := pipe.Exec(ctx)
	return err
}

func GetAllActiveVisits() ([]models.ActiveVisit, error) {
	ctx := context.Background()

	keys, err := rdb.Keys(ctx, "location:visit:*").Result()
	if err != nil {
		return nil, err
	}

	var visits []models.ActiveVisit

	for _, key := range keys {
		data, err := rdb.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var visit models.ActiveVisit
		if err := json.Unmarshal(data, &visit); err != nil {
			continue
		}

		visits = append(visits, visit)
	}

	return visits, nil
}
