package util

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"math/rand"
	"strconv"
	"time"
)

type Generator struct {
	Rand *rand.Rand
}

func NewGenerator() *Generator {
	return &Generator{
		Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Generator) RandomInt(min, max int) int {
	return min + g.Rand.Intn(max-min+1)
}

func (g *Generator) RandomInt32(min, max int32) int32 {
	return min + g.Rand.Int31n(max-min+1)
}

func (g *Generator) RandomInt64(min, max int64) int64 {
	return min + g.Rand.Int63n(max-min+1)
}

func (g *Generator) RandomName() string {
	return faker.FirstName()
}

func (g *Generator) RandomMoney() float64 {
	randomFloat := 1 + g.Rand.Float64()*(10000-1)
	formattedMoney := fmt.Sprintf("%.2f", randomFloat)
	if money, err := strconv.ParseFloat(formattedMoney, 64); err != nil {
		return 100
	} else {
		return money
	}
}

func (g *Generator) RandomCurrency() string {
	currencies := []string{"USD", "MYR"}
	n := len(currencies)
	return currencies[g.RandomInt(0, n-1)]
}
