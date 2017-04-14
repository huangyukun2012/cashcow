package utils
import (
	redis "gopkg.in/redis.v5"
	"time"
)
var redisclient *redis.Client

func InitRedis() {
	redisclient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}


func Set(url string, html string) error{
	err := redisclient.Set(url, html, 0).Err()
	return err
}

func Get(url string) (string,error) {
	val, err := redisclient.Get(url).Result()
	return val, err
}


func SetRedis(url string, html string) error{
	err := redisclient.Set(url, html, 3*time.Minute).Err()
	return err
}

func GetRedis(url string) (string,error) {
	val, err := redisclient.Get(url).Result()
	return val, err
}
