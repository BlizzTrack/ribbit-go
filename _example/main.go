package main

import (
	"log"
	"math/rand"
	"time"

	ribbit "github.com/blizztrack/ribbit-go"
)

func main() {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator

	red := ribbit.NewRibbitClient("us")

	sum, err := red.Summary()
	if err != nil {
		panic(err)
	}

	game := sum[rand.Intn(len(sum))]

	log.Println(game.Versions())
}
