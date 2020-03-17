package pkg

// "github.com/YOVO-LABS/theseus/db"

// import (
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"math"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/YOVO-LABS/theseus/db"
// 	"github.com/go-redis/redis/v7"
// 	"github.com/jinzhu/gorm"
// 	// "github.com/YOVO-LABS/theseus/db"
// 	// "github.com/go-redis/redis/v7"
// 	// "github.com/jinzhu/gorm"
// 	// _ "github.com/jinzhu/gorm/dialects/mysql"
// 	// "github.com/shopspring/decimal"
// 	// _ "github.com/shopspring/decimal"
// 	// "google.golang.org/grpc/codes"
// 	// "google.golang.org/grpc/status"
// )

// // //PrizeList ...
// // type PrizeList struct {
// // 	rank       int
// // 	count      int
// // 	totalScore int
// // }
// // ​
// // //LeaderData ...
// // type LeaderData struct {
// // 	score  int
// // 	member interface{}
// // }
// // ​
// // //RankData ...
// // type RankData struct {
// // 	rank       int
// // 	leaderData LeaderData
// // 	amount     float64
// // 	currency   string
// // }
// // ​
// // //LeaderBoard ..
// // type LeaderBoard struct {
// // 	name         string
// // 	game         string
// // 	currency     string
// // 	rankIndex    int
// // 	rankCutoff   int
// // 	prizeDefault []int
// // 	mongoClient  interface{}
// // }
// // ​
// type MvfList struct {
// 	ID        string `json:"id"`
// 	SID       string `json:"sId"`
// 	Exp       int64  `json:"exp"`
// 	CreatedAt string `json:"createdAt"`
// 	ChalName  string `json:"chalName"`
// 	State     string `json:"state"`
// 	Meta      string `json:"meta"`
// 	Start     int64  `json:"start"`
// }

// type sqlResult struct {
// 	Version string
// }

// // ConnectRedis creates a redis connection
// func ConnectRedis() (*redis.Client, error) {
// 	client := redis.NewClient(&redis.Options{
// 		Addr: "redis-17864.c1.ap-southeast-1-1.ec2.cloud.redislabs.com:17864",
// 		//Addr:     "10.115.202.164:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})
// 	pong, err := client.Ping().Result()
// 	fmt.Println(pong, err)
// 	return client, err
// }

// func ConnectSQL() (*gorm.DB, error) {
// 	var dbVersion sqlResult
// 	host := "34.93.3.207"
// 	database := "prod"
// 	// host := "10.65.96.3"
// 	// database := "prod"
// 	port := 3306
// 	username := "theseus_prod"
// 	password := "AK6xgnkDwqKh2HdH"

// 	dbSource := fmt.Sprintf(
// 		"%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True",
// 		username,
// 		password,
// 		host,
// 		port,
// 		database,
// 	)

// 	if flag.Lookup("test.v") != nil {
// 		dbSource = os.Getenv("dbSourceTest")
// 	}

// 	db, err := gorm.Open("mysql", dbSource)

// 	db.DB().SetConnMaxLifetime(time.Minute * 5)
// 	db.DB().SetMaxIdleConns(0)
// 	db.DB().SetMaxOpenConns(10)
// 	//	db.LogMode(true)

// 	if err := db.Raw("SELECT version() as version").Scan(&dbVersion).Error; err != nil {
// 		return db, err
// 	}

// 	fmt.Println("Sql version ", dbVersion)
// 	return db, err
// }

// func (l *LeaderBoard) calculateRank(rawLeaderData []LeaderData, cursor int) ([]*RankData, map[int]PrizeList) {
// 	rankList := []*RankData{}
// 	prizeList := make(map[int]PrizeList)

// 	idx := cursor

// 	currRank := 1

// 	for _, data := range rawLeaderData {
// 		tieRank := idx + 1
// 		if tieRank <= l.rankCutoff {
// 			pl, ok := prizeList[data.score]
// 			if ok == false {
// 				prizeList[data.score] = PrizeList{rank: currRank, count: 1, totalScore: l.prizeDefault[l.rankIndex+1]}
// 				currRank++

// 			} else {
// 				pl.totalScore += l.prizeDefault[l.rankIndex+1]
// 				pl.count++
// 				prizeList[data.score] = pl
// 			}

// 			if l.rankCutoff <= idx {
// 				pl.totalScore += 0
// 			} else {
// 				if (idx + 1 - l.prizeDefault[l.rankIndex]) == 0 {
// 					l.rankIndex += 2
// 				}
// 			}
// 		}
// 		rankList = append(rankList, &RankData{rank: currRank - 1, leaderData: data})
// 		idx++
// 	}

// 	return rankList, prizeList
// }

// //CompetitionRanking ...
// type CompetitionRanking struct {
// 	leaderBoard LeaderBoard
// }

// //NewCompetitionRanking ...
// func NewCompetitionRanking(lb *LeaderBoard) *CompetitionRanking {
// 	return &CompetitionRanking{leaderBoard: *lb}
// }

// func (c *CompetitionRanking) calculatePrizeDistribution(rl []*RankData, pl map[int]PrizeList) []*RankData {
// 	for idx, member := range rl {
// 		if val, ok := pl[member.leaderData.score]; ok {
// 			normalizedAmount := float64(val.totalScore / val.count)
// 			rl[idx].amount, rl[idx].currency = math.Round(normalizedAmount*100)/100, c.leaderBoard.currency

// 		} else {
// 			rl[idx].amount, rl[idx].currency = 0, ""
// 		}
// 	}
// 	return rl
// }

// //FetchCalculatedScore
// func (c *CompetitionRanking) FetchCalculatedScore(key string, prizeStr string, redisClient *redis.Client) []*RankData {
// 	var prizeDistInt []int
// 	prizeListStr := strings.Split(prizeStr, ",")
// 	for _, prize := range prizeListStr {
// 		i1, err := strconv.Atoi(prize)
// 		if err == nil {
// 			//fmt.Println(err)
// 		}
// 		prizeDistInt = append(prizeDistInt, i1)
// 	}

// 	leaderData := []LeaderData{}
// 	c.leaderBoard.name = "Test"
// 	c.leaderBoard.prizeDefault = prizeDistInt
// 	c.leaderBoard.rankCutoff = 1000
// 	c.leaderBoard.rankIndex = 0
// 	c.leaderBoard.game = "Test"
// 	c.leaderBoard.currency = "Test"
// 	response, err := redisClient.ZRevRangeWithScores(key, 0, -1).Result()
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	for _, res := range response {
// 		leaderData = append(leaderData, LeaderData{score: int(res.Score), member: res.Member})
// 	}

// 	rankList, prizeList := c.leaderBoard.calculateRank(leaderData, 0)
// 	rl := c.calculatePrizeDistribution(rankList, prizeList)
// 	return rl
// }

func main() {
	// var competition CompetitionRanking
	// DB, err := ConnectSQL()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer DB.Close()
	// redisClient, err := ConnectRedis()

	// var challenges []db.Challenge

	// if err := DB.Preload("ChallengeRounds", func(db *gorm.DB) *gorm.DB {
	// 	return db.Order("ChallengeRounds.level ASC")
	// }).Where("Challenges.state = ? and Challenges.expiry < ?", "LIVE", time.Now().Unix()).Order("Challenges.createdAt desc").Find(&challenges).Error; err != nil || len(challenges) < 1 {
	// 	fmt.Println("NO Challenges")
	// }

	// for _, chal := range challenges {
	// 	key := fmt.Sprintf("ch_%v_%v_votes", chal.ID, chal.RoundPostLevel)
	// 	rankList := competition.FetchCalculatedScore(key, chal.ChallengeRounds[0].Winner, redisClient)
	// 	if len(rankList) < 1 || rankList == nil {
	// 		continue
	// 	}
	// 	for _, rank := range rankList {
	// 		// var post db.Post
	// 		if err := DB.Where("sid=? and username=?",
	// 			fmt.Sprintf("%v_%v", chal.ID, chal.RoundPostLevel), rank.leaderData.member).Find(&post).Error; err != nil {
	// 			fmt.Println("Not found", chal.ID, chal.RoundPostLevel, rank.leaderData.member)
	// 			continue
	// 		}
	// 		meta := fmt.Sprintf("%v|%v|%v", rank.rank, rank.amount, rank.leaderData.score)
	// 		post.Meta = meta
	// 		// DB.Save(&post)
	// 		// fmt.Println(post.ID, meta)
	// 		// err = updateWallet(DB, &post, rank.amount, meta, redisClient)
	// 		// if err != nil {
	// 		// 	fmt.Println("Wallet ", err)
	// 		// }
	// 		// err = updateChallengeInCache(&post, meta, redisClient)
	// 		// if err != nil {
	// 		// 	fmt.Println("challenge cache ", err)
	// 		// }
	// 		//move leaderboard to cloud storage
	// 		time.Sleep(1 * time.Second)
	// 	}
	// 	//EXP challenge
	// 	// chalUpdate := DB.Table("Challenges").Where("id = ?", chal.ID).UpdateColumn("state", "DONE")
	// 	// if chalUpdate.RowsAffected == 0 || chalUpdate.Error != nil {
	// 	// }
	// 	time.Sleep(5 * time.Second)
	// }
}

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

// // func updateWallet(dbConn *gorm.DB, post *db.Post, amount float64, meta string, redisClient *redis.Client) error {
// // 	tx := dbConn.Begin()
// // 	defer func() {
// // 		if r := recover(); r != nil {
// // 			tx.Rollback()
// // 		}
// // 	}()​
// // 	if err := tx.Error; err != nil {
// // 		return err
// // 	}

// // 	add := tx.Model(&db.User{}).Where("id=?", post.UserID).UpdateColumn("Inr", gorm.Expr("inr + ? ", decimal.NewFromInt(amount)))
// // 	if add.RowsAffected == 0 {
// // 		tx.Rollback()
// // 		return add.Error
// // 	} else if add.Error != nil {
// // 		tx.Rollback()
// // 		return add.Error
// // 	}

// // 	transaction := db.Transaction{
// // 		App:      "yovo",
// // 		Currency: "INR",
// // 		UserID:   post.UserID,
// // 		Amount:   amount,
// // 		Meta:     fmt.Sprintf("%v|%v", post.ID, meta),
// // 		Type:     "CR",
// // 		Category: "CHAL",
// // 		IP:       "10.0.0.1",
// // 		State:    "DONE",
// // 	}
// // 	fmt.Println(transaction)

// // 	if err := tx.Model("Transactions").Create(&transaction).Error; err != nil {
// // 		tx.Rollback()
// // 		return status.Error(codes.Internal, "Internal Error. Contact Support")
// // 	}

// // 	//update wallet entry in redis
// // 	_, err := redisClient.HIncrBy(fmt.Sprintf("u:%v", post.UserID), "inr", amount).Result()
// // 	if err != nil {
// // 		return err
// // 	}

// // 	return tx.Commit().Error
// // }
