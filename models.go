package main

type Player struct {
	ID          string `bson:"_id,omitempty"`
	TelegramID  string `bson:"telegramId"`
	Name        string `bson:"name"`
	Points      int    `bson:"points"`
	Energy      int    `bson:"energy"`
	LevelEnergy int    `bson:"levelEnergy"`
	LevelX10    int    `bson:"levelX10"`
	LevelX100   int    `bson:"levelX100"`
	LevelX1000  int    `bson:"levelX1000"`
}
