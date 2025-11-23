package model

import "time"

type MessageTag struct {
	ID        string         `json:"id" bson:"_id"` // or appropriate type (e.g. primitive.ObjectID)
	MessageID string         `json:"message_id" bson:"message_id"`
	Tags      []TagWithScore `json:"tags" bson:"tags"`
	CreatedAt time.Time      `json:"created_at" bson:"created_at"`
}

type Tag struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

type TagWithScore struct {
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Score       float64 `json:"score" bson:"score"`
}
