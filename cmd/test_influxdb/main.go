package main

import (
	"fmt"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"log"
	"time"
)


func main() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
	})
	if err != nil {
		log.Fatal("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	q := client.NewQuery("SELECT * FROM tension", "oxygen73lab", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		for _, x := range response.Results{
			for _,row := range x.Series{
				fmt.Printf("%+v\n", row.Columns)
				for _, v := range row.Values{
					t,err := time.Parse(time.RFC3339, v[0].(string))
					if err != nil {
						panic(err)
					}
					strTime :=  t.Format("2006-01-02 15:04:05.000")
					fmt.Printf("%s ", strTime)
					for i,col := range row.Columns{
						if i == 0 {
							continue
						}
						fmt.Printf("%s=%v", col, v[i])
						if i == len(row.Columns) - 1{
							fmt.Print("\n")
						} else {
							fmt.Print(", ")
						}
					}
				}

			}

		}

	}
}
