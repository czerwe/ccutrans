package ccuprocessing

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/influxdata/influxdb/client/v2"
	// "os"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Klimaresponse struct {
	Status      bool      `json:"status"`
	Time        time.Time `json:"time"`
	Value       float64   `json:"value"`
	Room        string    `json:"room"`
	Measurement string    `json:"measurement"`
}

func QueryHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	measurement := strings.ToLower(vars["measurement"])

	room := strings.ToLower(vars["room"])

	log.WithFields(log.Fields{
		"measurement": measurement,
		"room":        room,
	}).Debug("Received query")

	url := fmt.Sprintf("http://10.0.0.4:8086")

	influxclient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: url,
	})
	defer influxclient.Close()

	query := fmt.Sprintf("select value from %v where \"room\" = '%v' order by time desc limit 1", measurement, room)

	res, err := QueryDB(influxclient, query, "homeatic")
	if err != nil {
		log.Fatal(err)
	}

	if len(res[0].Series) == 1 {
		for _, row := range res[0].Series[0].Values {

			parsedNumber, parseError := strconv.ParseFloat(string(row[1].(json.Number)), 64)

			// fmt.Printf("%T\n", parseError)
			if parseError != nil {
				log.Fatal(parseError)
			}

			timeString, timeError := time.Parse(time.RFC3339, row[0].(string))
			if timeError != nil {
				log.Fatal(timeError)
			}

			log.WithFields(log.Fields{
				"time":  timeString,
				"value": parsedNumber,
			}).Info("query answer")

			response := Klimaresponse{
				Status:      true,
				Time:        timeString,
				Value:       parsedNumber,
				Room:        room,
				Measurement: measurement}

			responseData, _ := json.Marshal(response)

			resp.Header().Set("Content-Type", "application/json")
			resp.Write(responseData)
		}
	} else {

		log.WithFields(log.Fields{
			"measurement": measurement,
			"room":        room,
		}).Info("No answers for query")

		response := Klimaresponse{
			Status:      false,
			Room:        room,
			Measurement: measurement}

		responseData, _ := json.Marshal(response)

		resp.Header().Set("Content-Type", "application/json")
		resp.Write(responseData)
	}

}
