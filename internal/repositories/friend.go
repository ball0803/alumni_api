package repositories

import (
	"alumni_api/internal/models"
	"alumni_api/internal/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func GetUserFriendByID(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (:UserProfile {user_id: $id})-[:FRIEND]->(f:UserProfile)
    RETURN collect({
        user_id: f.user_id,
        username: f.username,
        first_name: f.first_name,
        last_name: f.last_name,
        first_name_eng: f.first_name_eng,
        last_name_eng: f.last_name_eng,
        profile_picture: f.profile_picture
    }) AS friends
  `

	result, err := session.Run(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	records, err := result.Single(ctx)
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	var friends []map[string]interface{}
	friendRecords, ok := records.Get("friends")
	if !ok {
		logger.Warn("No friends found for user")
		return nil, fiber.NewError(fiber.StatusNotFound, "No friends found for this user")
	}
	friendList, ok := friendRecords.([]interface{})
	if !ok {
		logger.Error("Failed to cast friends data to []interface{}")
		return nil, fiber.NewError(http.StatusInternalServerError, "Error processing friends data")
	}

	for _, friendData := range friendList {
		friendMap := friendData.(map[string]interface{})
		for key, value := range friendMap {
			if utils.IsEmpty(value) {
				delete(friendMap, key)
			}
		}
		friends = append(friends, friendMap)
	}

	return friends, nil
}

func AddFriend(ctx context.Context, driver neo4j.DriverWithContext, userID1 string, userID2 string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			query := `
        MATCH (u1:UserProfile {user_id: $userID1}), (u2:UserProfile {user_id: $userID2})
        MERGE (u1)-[r:FRIEND]->(u2)
        ON CREATE SET r.created_timestamp = timestamp()
        MERGE (u2)-[r2:FRIEND]->(u1)
        ON CREATE SET r2.created_timestamp = timestamp()
        RETURN u1, u2
      `

			result, err := tx.Run(ctx, query, map[string]interface{}{
				"userID1": userID1,
				"userID2": userID2,
			})

			if err != nil {
				logger.Error("Failed to add friend", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
			}

			_, err = result.Single(ctx)
			if err != nil {
				logger.Error("Failed to retrieve result", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve result after creating relationship")
			}
			return nil, nil
		})

	if err != nil {
		return err
	}

	return nil
}

func Unfriend(ctx context.Context, driver neo4j.DriverWithContext, userID1 string, userID2 string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	if userID1 == userID2 {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot unfriend oneself")
	}

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			query := `
      MATCH (u1:UserProfile {user_id: $userID1})-[r1:FRIEND]->(u2:UserProfile {user_id: $userID2})
      DELETE r1
      WITH u1, u2
      MATCH (u2)-[r2:FRIEND]->(u1)
      DELETE r2
      RETURN u1, u2
    `
			_, err := tx.Run(ctx, query, map[string]interface{}{
				"userID1": userID1,
				"userID2": userID2,
			})
			if err != nil {
				logger.Error("Failed to add friend", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
			}
			return nil, nil
		})

	if err != nil {
		return err
	}
	return nil
}

func GetFOAF(ctx context.Context, driver neo4j.DriverWithContext, user_id, other_id string, degree int, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j", // Replace with your database name if different
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// Dynamically construct the query with the degree value
	query := fmt.Sprintf(`
    MATCH (a:UserProfile {user_id: $user_id}), (b:UserProfile {user_id: $other_id})
    MATCH p = (a)-[:FRIEND*1..%d]-(b)
    WITH p, nodes(p) AS nodeList
    UNWIND range(1, size(nodeList) - 2) AS idx
    WITH nodeList[idx] AS n, idx, p
    WITH p, COLLECT({
      user_id: n.user_id,
      contact: {
        email: n.email,
        linkedin: n.linkedin,
        phone: n.phone,
        facebook: n.facebook
      },
      profile_picture: n.profile_picture,
      fullname: n.first_name + ' ' + n.last_name,
      fullname_eng: n.first_name_eng + ' ' + n.last_name_eng,
      depth: idx
    }) AS nodeInfoList
    WITH DISTINCT nodeInfoList
    RETURN nodeInfoList
    ORDER BY size(nodeInfoList) ASC
    LIMIT 50
  `, degree)

	params := map[string]interface{}{
		"user_id":  user_id,
		"other_id": other_id,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to execute Neo4j query", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to execute Neo4j query")
	}

	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect Neo4j records", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect Neo4j records")
	}

	var foaf []map[string]interface{}

	for _, record := range records {
		// Extract the nodeInfoList from the record
		nodeInfoList, ok := record.Get("nodeInfoList")
		if !ok {
			logger.Warn("Missing nodeInfoList in Neo4j record")
			continue
		}

		// Convert nodeInfoList to []map[string]interface{}
		nodeInfoListSlice, ok := nodeInfoList.([]interface{})
		if !ok {
			logger.Warn("Invalid nodeInfoList format in Neo4j record")
			continue
		}

		for _, nodeInfo := range nodeInfoListSlice {
			nodeInfoMap, ok := nodeInfo.(map[string]interface{})
			if !ok {
				logger.Warn("Invalid nodeInfo format in Neo4j record")
				continue
			}

			foaf = append(foaf, nodeInfoMap)
		}
	}

	return foaf, nil
}
