package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerResponse struct {
	Name            string  `json:"name"`
	Points          int     `json:"points"`
	Energy          int     `json:"energy"`
	MultiplierX10   float64 `json:"multiplierX10"`
	MultiplierX100  float64 `json:"multiplierX100"`
	MultiplierX1000 float64 `json:"multiplierX1000"`
}

type LeaderboardResponse struct {
	Name   string `json:"name"`
	Points int    `json:"points"`
}

type UpdateEnergyRequest struct {
	Energy int `json:"energy"`
}

type UpdatePointsRequest struct {
	Points int `json:"points"`
}

type UpdateSkillsRequest struct {
	LevelEnergy int `json:"levelEnergy"`
	LevelX10    int `json:"levelX10"`
	LevelX100   int `json:"levelX100"`
	LevelX1000  int `json:"levelX1000"`
}

type GetSkillsResponse struct {
	LevelEnergy       int     `json:"levelEnergy"`
	MaxEnergy         float32 `json:"maxEnergy"`
	NextMaxEnergy     float32 `json:"nextMaxEnergy"`
	CostUpgradeEnergy int     `json:"costUpgradeEnergy"`

	LevelX10             int     `json:"levelX10"`
	CurrentMultiplierX10 float32 `json:"currentMultiplierX10"`
	NextMultiplierX10    float32 `json:"nextMultiplierX10"`
	CostUpgradeX10       int     `json:"costUpgradeX10"`

	LevelX100             int     `json:"levelX100"`
	CurrentMultiplierX100 float32 `json:"currentMultiplierX100"`
	NextMultiplierX100    float32 `json:"nextMultiplierX100"`
	CostUpgradeX100       int     `json:"costUpgradeX100"`

	LevelX1000             int     `json:"levelX1000"`
	CurrentMultiplierX1000 float32 `json:"currentMultiplierX1000"`
	NextMultiplierX1000    float32 `json:"nextMultiplierX1000"`
	CostUpgradeX1000       int     `json:"costUpgradeX1000"`
}

func getPlayerHandler(w http.ResponseWriter, r *http.Request) {
	telegramID := r.URL.Query().Get("telegramId")
	log.Printf("Fetching player data for telegramId: %s", telegramID)

	var user Player
	collection := db.Collection("users")
	err := collection.FindOne(context.TODO(), bson.M{"telegramId": telegramID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("User not found for telegramId: %s", telegramID)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error finding user for telegramId:%s, error: %v", telegramID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := PlayerResponse{
		Name:            user.Name,
		Points:          user.Points,
		Energy:          user.Energy,
		MultiplierX10:   float64(user.LevelX10 * 10),
		MultiplierX100:  float64(user.LevelX100),
		MultiplierX1000: float64(user.LevelX1000 / 10),
	}

	writeJSONResponse(w, resp)
}

func updateEnergyHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateEnergyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	telegramId := vars["telegramId"]
	log.Printf("Updating energy for telegramId: %s, new energy: %d", telegramId, req.Energy)

	updateField(w, telegramId, "energy", req.Energy)
}

func updatePointsHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdatePointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	telegramId := vars["telegramId"]
	log.Printf("Updating points for telegramId: %s, new points: %d", telegramId, req.Points)

	updateField(w, telegramId, "points", req.Points)
}

func getSkillsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramId := vars["telegramId"]
	log.Printf("Fetching skills for telegramId: %s", telegramId)

	var user Player
	collection := db.Collection("users")
	err := collection.FindOne(context.TODO(), bson.M{"telegramId": telegramId}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("User not found for telegramId: %s", telegramId)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error finding user for telegramId:%s, error: %v", telegramId, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := calcSkills(user)
	writeJSONResponse(w, resp)
}

func updateSkillsHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateSkillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	telegramId := vars["telegramId"]
	log.Printf("Updating skills for telegramId: %s", telegramId)

	collection := db.Collection("users")
	filter := bson.M{"telegramId": telegramId}
	update := bson.M{
		"$set": bson.M{
			"levelEnergy": req.LevelEnergy,
			"levelX10":    req.LevelX10,
			"levelX100":   req.LevelX100,
			"levelX1000":  req.LevelX1000,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	var user Player
	err := collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&user)
	if err != nil {
		log.Printf("Error updating skills for telegramId: %s, error: %v", telegramId, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := calcSkills(user)
	writeJSONResponse(w, resp)
}

func getLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetching leaderboard")
	collection := db.Collection("users")
	options := options.Find().SetSort(bson.D{{Key: "points", Value: -1}}).SetLimit(10)

	cursor, err := collection.Find(context.TODO(), bson.M{}, options)
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var leaderboard []LeaderboardResponse
	for cursor.Next(context.TODO()) {
		var user Player
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Error decoding user: %v", err)
			continue
		}
		leaderboard = append(leaderboard, LeaderboardResponse{
			Name:   user.Name,
			Points: user.Points,
		})
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, leaderboard)
}

func calcSkills(user Player) GetSkillsResponse {
	return GetSkillsResponse{
		LevelEnergy:            user.LevelEnergy,
		MaxEnergy:              calcSkillStats(user.LevelEnergy, 10),
		NextMaxEnergy:          calcSkillStats(user.LevelEnergy+1, 10),
		CostUpgradeEnergy:      calcUpgradeCost(user.LevelEnergy, 1.5, 10),
		LevelX10:               user.LevelX10,
		CurrentMultiplierX10:   calcSkillStats(user.LevelX10, 2.5),
		NextMultiplierX10:      calcSkillStats(user.LevelX10+1, 2.5),
		CostUpgradeX10:         calcUpgradeCost(user.LevelX10, 1.2, 5),
		LevelX100:              user.LevelX100,
		CurrentMultiplierX100:  calcSkillStats(user.LevelX100, 0.3),
		NextMultiplierX100:     calcSkillStats(user.LevelX100+1, 0.3),
		CostUpgradeX100:        calcUpgradeCost(user.LevelX100, 1.3, 5),
		LevelX1000:             user.LevelX1000,
		CurrentMultiplierX1000: calcSkillStats(user.LevelX1000, 0.05),
		NextMultiplierX1000:    calcSkillStats(user.LevelX1000+1, 0.05),
		CostUpgradeX1000:       calcUpgradeCost(user.LevelX1000, 1.4, 5),
	}
}

func calcSkillStats(level int, factor float32) float32 {
	return float32(level) * factor
}

func calcUpgradeCost(level int, base, multiplier float64) int {
	return int(multiplier * math.Pow(base, float64(level-1)))
}

func updateField(w http.ResponseWriter, telegramId string, field string, value interface{}) {
	collection := db.Collection("users")
	filter := bson.M{"telegramId": telegramId}
	update := bson.M{"$set": bson.M{field: value}}

	_, err := collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("Error updating %s for telegramId: %s, error: %v", field, telegramId, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Updated %s successfully for telegramId: %s", field, telegramId)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
