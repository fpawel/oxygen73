package data

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestAddProductVoltages(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for n := 0; n < 100; n++{
		for place := 0; place < 50; place++{
			AddProductVoltages([]ProductVoltage{{
				Time:    time.Now(),
				Place:   place,
				Number:  place + 100,
				Tension: rand.Float64() * 100,
			}})
		}
		time.Sleep(time.Millisecond)
		AddAmbient(Ambient{
			Time:        time.Now(),
			Temperature: rand.Float64() * 100,
			Pressure:    rand.Float64() * 100,
			Humidity:    rand.Float64() * 100,
		})
		SaveAndCleanCache()
	}
}

func TestUpdatedAt(t *testing.T) {
	fmt.Println(UpdatedAt())
}

