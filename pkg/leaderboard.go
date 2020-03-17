package pkg

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	// "github.com/YOVO-LABS/theseus/db"
	"github.com/go-redis/redis/v7"

	//_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/shopspring/decimal"
)

//RankInfo denotes no. of participants with same rank and their cumulative score
type RankInfo struct {
	rank       int
	count      int
	totalScore int
}

//LeaderData ...
type LeaderData struct {
	score  int
	member interface{}
}

//RankBoard ...
type RankBoard struct {
	rank       int
	leaderData LeaderData
	amount     float64
	currency   string
}

//LeaderBoard ..
type LeaderBoard struct {
	name         string
	game         string
	currency     string
	rankIndex    int
	rankCutoff   int
	prizeDefault []int
	mongoClient  interface{}
}

// NewLeaderBoard ...
func NewLeaderBoard() *LeaderBoard {
	return &LeaderBoard{}
}

//MvfList ...
type MvfList struct {
	ID        string `json:"id"`
	SID       string `json:"sId"`
	Exp       int64  `json:"exp"`
	CreatedAt string `json:"createdAt"`
	ChalName  string `json:"chalName"`
	State     string `json:"state"`
	Meta      string `json:"meta"`
	Start     int64  `json:"start"`
}

type sqlResult struct {
	Version string
}

//calculateRank fetches RankBoard & RankInfo based on raw leaderboard data coming from redis
func (l *LeaderBoard) calculateRank(rawLeaderData []LeaderData, cursor int) ([]*RankBoard, map[int]RankInfo) {
	rb := []*RankBoard{}
	ri := make(map[int]RankInfo)
	idx := cursor
	currRank := 1
	for _, data := range rawLeaderData {
		if idx+1 <= l.rankCutoff {
			val, ok := ri[data.score]
			if ok == false {
				ri[data.score] = RankInfo{rank: currRank, count: 1, totalScore: l.prizeDefault[l.rankIndex+1]}
				currRank++
			} else {
				val.totalScore += l.prizeDefault[l.rankIndex+1]
				val.count++
				ri[data.score] = val
			}
			if l.rankCutoff <= idx {
				val.totalScore += 0
			} else {
				if (idx + 1 - l.prizeDefault[l.rankIndex]) == 0 {
					l.rankIndex += 2
				}
			}
		}
		rb = append(rb, &RankBoard{rank: currRank - 1, leaderData: data})
		idx++
	}
	return rb, ri
}

//CompetitionRanking ...
type CompetitionRanking struct {
	leaderBoard LeaderBoard
}

//NewCompetitionRanking ...
func NewCompetitionRanking(lb *LeaderBoard) *CompetitionRanking {
	return &CompetitionRanking{leaderBoard: *lb}
}

func (c *CompetitionRanking) calculatePrizeDistribution(rl []*RankBoard, pl map[int]RankInfo) []*RankBoard {
	for idx, member := range rl {
		if val, ok := pl[member.leaderData.score]; ok {
			normalizedAmount := float64(val.totalScore / val.count)
			rl[idx].amount, rl[idx].currency = math.Round(normalizedAmount*100)/100, c.leaderBoard.currency
		} else {
			rl[idx].amount, rl[idx].currency = 0, ""
		}
	}
	return rl
}

//FetchLeaderboardScore ...
func (c *CompetitionRanking) FetchLeaderboardScore(key string, prizeStr string, redisClient *redis.Client) []*RankBoard {
	var prizeDistInt []int
	prizeListStr := strings.Split(prizeStr, ",")

	for _, prize := range prizeListStr {
		i1, err := strconv.Atoi(prize)
		if err == nil {
		}
		prizeDistInt = append(prizeDistInt, i1)
	}

	leaderData := []LeaderData{}
	c.leaderBoard.name = "Test"
	c.leaderBoard.prizeDefault = prizeDistInt
	c.leaderBoard.rankCutoff = 1000
	c.leaderBoard.rankIndex = 0
	c.leaderBoard.game = "Test"
	c.leaderBoard.currency = "Test"
	response, err := redisClient.ZRevRangeWithScores(key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, res := range response {
		leaderData = append(leaderData, LeaderData{score: int(res.Score), member: res.Member})
	}
	rankBoard, rankInfo := c.leaderBoard.calculateRank(leaderData, 0)
	rl := c.calculatePrizeDistribution(rankBoard, rankInfo)
	return rl
}
