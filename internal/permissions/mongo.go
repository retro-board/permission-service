package permissions

import (
	"context"
	"errors"
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/permission-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Config *config.Config
	CTX    context.Context
}

func NewMongo(c *config.Config) *Mongo {
	return &Mongo{
		Config: c,
		CTX:    context.Background(),
	}
}

type DataSet struct {
	UserID      string       `bson:"user_id" json:"user_id"`
	Generated   int64        `bson:"generated" json:"generated"`
	Permissions []Permission `bson:"permissions" json:"permissions"`
}

func (m *Mongo) getConnection() (*mongo.Client, error) {
	if m.Config.Local.Development {
		client, err := mongo.Connect(m.CTX, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", m.Config.Mongo.Host)))
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	client, err := mongo.Connect(
		m.CTX,
		options.Client().ApplyURI(fmt.Sprintf(
			"mongodb+srv://%s:%s@%s",
			m.Config.Mongo.Username,
			m.Config.Mongo.Password,
			m.Config.Mongo.Host)),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (m *Mongo) Get(userID string) (*DataSet, error) {
	client, err := m.getConnection()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	var dataSet DataSet
	err = client.
		Database("permissions").
		Collection("permissions").
		FindOne(m.CTX, bson.M{"user_id": userID}).
		Decode(&dataSet)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &dataSet, nil
}

func (m *Mongo) Create(data DataSet) error {
	client, err := m.getConnection()
	if err != nil {
		return err
	}

	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	_, err = client.
		Database("permissions").
		Collection("permissions").
		InsertOne(m.CTX, data)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongo) Update(data DataSet) error {
	client, err := m.getConnection()
	if err != nil {
		return err
	}

	defer func() {
		if err := client.Disconnect(m.CTX); err != nil {
			bugLog.Info(err)
		}
	}()

	_, err = client.
		Database("permissions").
		Collection("permissions").
		UpdateOne(m.CTX, bson.M{"user_id": data.UserID}, bson.M{"$set": data})
	if err != nil {
		return err
	}

	return nil
}
