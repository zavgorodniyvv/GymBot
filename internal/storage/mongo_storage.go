package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Session и UserData оставляем те же структуры, что были (можно вынести их в отдельный файл models.go)
type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// Создаём подключение
func NewMongoStorage(uri string) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	db := client.Database("gymbot") // имя БД из connection string
	coll := db.Collection("users")  // коллекция, где будут лежать UserData

	return &MongoStorage{
		client:     client,
		collection: coll,
	}, nil
}

// Сохраняем пользователя
func (m *MongoStorage) SaveUser(u *UserData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// upsert (обновить или вставить, если нет)
	_, err := m.collection.UpdateOne(
		ctx,
		bson.M{"user_id": u.UserId},
		bson.M{"$set": u},
		options.Update().SetUpsert(true),
	)
	return err
}

// Загружаем пользователя
func (m *MongoStorage) LoadUser(userID int64) (*UserData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u UserData
	err := m.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return NewUser(userID), nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Завершаем тренировку (похож на файловую версию)
func (m *MongoStorage) FinishWorkout(u *UserData) (Session, error) {
	if len(u.CurrentWorkout) == 0 {
		return Session{}, ErrEmptyWorkout
	}
	maxSet, total := 0, 0
	for _, r := range u.CurrentWorkout {
		if r > maxSet {
			maxSet = r
		}
		total += r
	}
	s := Session{
		Date:       time.Now(),
		Sets:       append([]int(nil), u.CurrentWorkout...),
		MaxSet:     maxSet,
		TotalReps:  total,
		IsFinished: true,
		Planned:    append([]int(nil), u.LastPlan...),
	}
	u.Sessions = append(u.Sessions, s)
	u.CurrentWorkout = nil
	if err := m.SaveUser(u); err != nil {
		return Session{}, err
	}
	return s, nil
}
