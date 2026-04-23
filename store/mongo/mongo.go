// Package mongo implements store.GameDB using MongoDB.
package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/21ess/frozeel/domain/game"
	"github.com/21ess/frozeel/provider"
	"github.com/21ess/frozeel/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	dbName             = "frozeel"
	collectionsColl    = "collections"
	countersColl       = "counters"
	collectionsCounter = "collection_id"
)

// MongoGameDB implements store.GameDB backed by MongoDB.
type MongoGameDB struct {
	client *mongo.Client
	db     *mongo.Database
}

// Verify interface compliance at compile time.
var _ store.GameDB = (*MongoGameDB)(nil)

// NewMongoGameDB connects to MongoDB and returns a GameDB implementation.
func NewMongoGameDB(ctx context.Context, uri string) (*MongoGameDB, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	db := client.Database(dbName)
	return &MongoGameDB{client: client, db: db}, nil
}

// Close disconnects from MongoDB.
func (m *MongoGameDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// EnsureIndexes creates necessary indexes. Call once at startup.
func (m *MongoGameDB) EnsureIndexes(ctx context.Context) error {
	coll := m.db.Collection(collectionsColl)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "popularity", Value: -1}},
		},
	})
	if err != nil {
		return fmt.Errorf("ensure indexes: %w", err)
	}
	return nil
}

// nextID atomically increments and returns the next auto-increment ID for the given sequence name.
func (m *MongoGameDB) nextID(ctx context.Context, seqName string) (int64, error) {
	coll := m.db.Collection(countersColl)
	filter := bson.M{"_id": seqName}
	update := bson.M{"$inc": bson.M{"seq": int64(1)}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int64 `bson:"seq"`
	}
	err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("next id for %q: %w", seqName, err)
	}
	return result.Seq, nil
}

func (m *MongoGameDB) GetCollection(ctx context.Context, id int64) (*game.Collection, error) {
	coll := m.db.Collection(collectionsColl)
	filter := bson.M{"id": id}

	var col game.Collection
	if err := coll.FindOne(ctx, filter).Decode(&col); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("collection %d not found", id)
		}
		return nil, fmt.Errorf("get collection: %w", err)
	}
	return &col, nil
}

func (m *MongoGameDB) ListCollections(ctx context.Context, count int) ([]*game.Collection, error) {
	coll := m.db.Collection(collectionsColl)
	opts := options.Find().
		SetSort(bson.D{{Key: "popularity", Value: -1}}).
		SetLimit(int64(count))

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*game.Collection
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode collections: %w", err)
	}
	return results, nil
}

func (m *MongoGameDB) CreateCollection(ctx context.Context, collection *game.Collection) error {
	id, err := m.nextID(ctx, collectionsCounter)
	if err != nil {
		return err
	}
	collection.ID = id

	now := time.Now()
	collection.CreatedAt = now
	collection.UpdatedAt = now

	coll := m.db.Collection(collectionsColl)
	if _, err := coll.InsertOne(ctx, collection); err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	return nil
}

func (m *MongoGameDB) BatchAddSubjectToCollection(ctx context.Context, subjects []*provider.Subject) error {
	if len(subjects) == 0 {
		return nil
	}

	coll := m.db.Collection(collectionsColl)

	docs := make([]interface{}, len(subjects))
	for i, s := range subjects {
		docs[i] = s
	}

	if _, err := coll.InsertMany(ctx, docs); err != nil {
		return fmt.Errorf("batch add subjects: %w", err)
	}
	return nil
}

func (m *MongoGameDB) IncrCollectionPopularity(ctx context.Context, collectionID int64) error {
	coll := m.db.Collection(collectionsColl)
	filter := bson.M{"id": collectionID}
	update := bson.M{
		"$inc": bson.M{"popularity": int64(1)},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("incr popularity: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("collection %d not found", collectionID)
	}
	return nil
}
