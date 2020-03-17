package pkg

import (
	"fmt"

	"github.com/go-redis/redis/v7"
)

// ConnectRedis creates a redis connection
func ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "redis-17864.c1.ap-southeast-1-1.ec2.cloud.redislabs.com:17864",
		//Addr:     "10.115.202.164:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client, err
}
