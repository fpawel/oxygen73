package data

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestAddProductVoltages(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for n := 0; n < 100; n++ {
		for place := 0; place < 50; place++ {
			AddProductVoltage(place, place+100, rand.Float64()*100)
		}
		AddAmbient(rand.Float64()*100, rand.Float64()*100, rand.Float64()*100)
		Save()
	}
}

func TestUpdatedAt(t *testing.T) {
	fmt.Println(lastSavedProductVoltage())
}
