package leaderboard

import (
	"context"

	// "github.com/YOVO-LABS/theseus/db"

	"go.uber.org/cadence/activity"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	calculateLeaderBoardName = "calculateLeaderBoard"
	pushLeaderboardScoreName = "pushLeaderboardScore"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		calculateLeaderBoard,
		activity.RegisterOptions{Name: calculateLeaderBoardName},
	)
}

type challengeInfo struct {
	// post     *db.Post
	amount   float64
	metaData string
}

func calculateLeaderBoard(ctx context.Context, jobID string) (map[int][]*challengeInfo, error) {
	return nil, nil
	// var competition pkg.CompetitionRanking
	// var chPost map[int][]*challengeInfo
	// var appConfig config.AppConfig
	// var kafkaClient adapter.KafkaAdapter

	// appConfig.Setup()
	// kafkaClient.Setup(&appConfig.Kafka)

	// DB, err := ConnectSQL()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer DB.Close()

	// redisClient, err := pkg.ConnectRedis()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer redisClient.Close()

	// var challenges []db.Challenge
	// if err := DB.Preload("ChallengeRounds", func(db *gorm.DB) *gorm.DB {
	// 	return db.Order("ChallengeRounds.level ASC")
	// }).Where("Challenges.state = ? and Challenges.expiry < ?", "LIVE", time.Now().Unix()).Order("Challenges.createdAt desc").Find(&challenges).Error; err != nil || len(challenges) < 1 {
	// 	fmt.Println("NO Challenges")
	// }
	// for _, chal := range challenges {
	// 	key := fmt.Sprintf("ch_%v_%v_votes", chal.ID, chal.RoundPostLevel)
	// 	rankBoard := competition.FetchCalculatedScore(key, chal.ChallengeRounds[0].Winner, redisClient)
	// 	if len(rankBoard) < 1 || rankBoard == nil {
	// 		continue
	// 	}
	// 	for _, rank := range rankBoard {
	// 		var post db.Post
	// 		if err := DB.Where("sid=? and username=?",
	// 			fmt.Sprintf("%v_%v", chal.ID, chal.RoundPostLevel), rank.leaderData.member).Find(&post).Error; err != nil {
	// 			fmt.Println("Not found", chal.ID, chal.RoundPostLevel, rank.leaderData.member)
	// 			continue
	// 		}
	// 		meta := fmt.Sprintf("%v|%v|%v", rank.rank, rank.amount, rank.leaderData.score)
	// 		post.Meta = meta
	// 		DB.Save(chInfo.post)
	// 		fmt.Println(post.ID, meta)
	// 		chInfo := challengeInfo{post: post, amount: rank.amount, metaData: meta}

	// 		msg, err := json.Marshal(chInfo)
	// 		if err != nil {
	// 			panic(err)
	// 			return
	// 		}

	// 		fmt.Println(string(msg))
	// 		err = kafkaClient.Producer.Publish(context.Background(), jobID, msg)
	// 		if err != nil {
	// 			panic(err)
	// 			return
	// 		}
	// 	}
	// 	// EXP challenge
	// }
	// return chPost, nil
}

// func pushLeaderboardScore(ctx context.Context, chPost map[int][]*challengeInfo) error {
// 	logger := activity.GetLogger(ctx)
// 	logger.Info("pushing learderboard score")

// 	redisClient, err := pkg.ConnectRedis()
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	defer redisClient.Close()

// 	DB, err := ConnectSQL()
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	defer DB.Close()

// 	for key, value := range chPost {
// 		for _, chInfo := range chPost[key] {
// 			DB.Save(chInfo.post)

// 			err = updateWallet(DB, ch.post, ch.amount, meta, redisClient)
// 			if err != nil {
// 				fmt.Println("Wallet ", err)
// 			}

// 			err = updateChallengeInCache(&post, meta, redisClient)
// 			if err != nil {
// 				fmt.Println("challenge cache ", err)
// 			}
// 			//move leaderboard to cloud storage
// 			time.Sleep(1 * time.Second)
// 		}
// 		chalUpdate := DB.Table("Challenges").Where("id = ?", chal.ID).UpdateColumn("state", "DONE")
// 		if chalUpdate.RowsAffected == 0 || chalUpdate.Error != nil {
// 		}
// 		time.Sleep(5 * time.Second)
// 	}
// 	return nil
// }

// func updateChallengeInCache(post *db.Post, meta string, redisClient *redis.Client) error {
// 	var mvfList []*MvfList
// 	userKey := fmt.Sprintf("u:%v", post.UserID)
// 	userData, err := redisClient.HMGet(userKey, "challenges").Result()
// 	if err != nil {
// 		return err
// 	}
// 	mvfStatus := userData[0]
// 	if mvfStatus != nil {
// 		err := json.Unmarshal([]byte(mvfStatus.(string)), &mvfList)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	for _, mvf := range mvfList {
// 		if mvf.ID == post.ID {
// 			mvf.Meta = meta
// 		}
// 	}
// 	jsonByte, _ := json.Marshal(mvfList)
// 	var m = make(map[string]interface{})
// 	m["challenges"] = string(jsonByte)
// 	_, err = redisClient.HMSet(userKey, m).Result()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func updateWallet(dbConn *gorm.DB, post *db.Post, amount float64, meta string, redisClient *redis.Client) error {
// 	tx := dbConn.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		}
// 	}()
// 	if err := tx.Error; err != nil {
// 		return err
// 	}
// 	add := tx.Model(&db.User{}).Where("id=?", post.UserID).UpdateColumn("Inr", gorm.Expr("inr + ? ", decimal.NewFromInt(amount)))
// 	if add.RowsAffected == 0 {
// 		tx.Rollback()
// 		return add.Error
// 	} else if add.Error != nil {
// 		tx.Rollback()
// 		return add.Error
// 	}
// 	transaction := db.Transaction{
// 		App:      "yovo",
// 		Currency: "INR",
// 		UserID:   post.UserID,
// 		Amount:   amount,
// 		Meta:     fmt.Sprintf("%v|%v", post.ID, meta),
// 		Type:     "CR",
// 		Category: "CHAL",
// 		IP:       "10.0.0.1",
// 		State:    "DONE",
// 	}
// 	fmt.Println(transaction)
// 	if err := tx.Model("Transactions").Create(&transaction).Error; err != nil {
// 		tx.Rollback()
// 		return status.Error(codes.Internal, "Internal Error. Contact Support")
// 	}
// 	//update wallet entry in redis
// 	_, err := redisClient.HIncrBy(fmt.Sprintf("u:%v", post.UserID), "inr", amount).Result()
// 	if err != nil {
// 		return err
// 	}
// 	return tx.Commit().Error
// }
