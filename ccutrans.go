// Transfer script for kilma metrics for the homeatic ccu system
//
// test urls:
// 	curl http://localhost:4040/temperature/BidCos-RF.LEQ0122563:1.TEMPERATURE/27.820000
// 	curl http://localhost:4040/temperature/BidCos-RF.LEQ0122563:1.TEMPERATURE/27.820000
package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	Influxhost  []string `short:"i" long:"host" required:"True" description:"Hostname or IP of Influxdb application"`
	Influxport  []int    `short:"p" long:"port" default:"8086" description:"Port of Influxdb application"`
	Mappingfile []string `short:"m" long:"mappingfile" default:"mapping.json" description:"Mapping file which provides serial/tag assigment"`
	Influxdb    []string `short:"d" long:"database" default:"homeatic" description:"The database name"`
	Loglevel    []string `long:"loglevel" default:"info" description:"loglevel" choice:"warn" choice:"info" choice:"debug"`
	Version     []bool   `long:"version" short:"v" default:"info" description:"show version" choice:false`
}

type influxmessage struct {
	measurement string
	tags        map[string]string
	fields      map[string]interface{}
	time        time.Time
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

var (
	msgChannel   chan influxmessage
	lastValues   map[string]float64
	mappingTable map[string][]map[string]string
	opts         Options
	level        log.Level
	version      string = "0.1.0"
)

func printVersion() {
	fmt.Printf("v%v\n", version)
}

func newInfluxmessage(measurement string, tags map[string]string, fields map[string]interface{}) influxmessage {
	// Creates an skeleton of an influxmessage
	// the influxmessage is used to wrap the point, meassurement and value to send it to
	// subroutine
	retVal := influxmessage{}
	retVal.measurement = measurement
	retVal.tags = tags
	retVal.fields = fields
	retVal.time = time.Now()
	return retVal
}

func readMappingFile(filename string) map[string][]map[string]string {
	//
	// Read a mapping file that assign tags to the serials
	//
	// The Syntax within the file is like:
	// {
	//    "leq0122563": [
	//         {
	//             "name": "room",
	//             "value": "dachgeschoss"
	//         }
	//    	]
	// }

	mappingFile, err := os.Open(filename)
	returnValue := make(map[string][]map[string]string)

	if err != nil {
		log.WithFields(log.Fields{
			"file":  filename,
			"error": err,
		}).Warning("Mapping file not readable")
	} else {
		defer mappingFile.Close()
		decoderJson := json.NewDecoder(mappingFile)
		if err := decoderJson.Decode(&returnValue); err != nil {
			fmt.Println(err)
		}
	}

	return returnValue
}

func main() {

	// version = "v1.0.0"

	_, err := flags.Parse(&opts)

	if err != nil {
		panic(err)
		os.Exit(1)
	}

	fmt.Println(opts)

	// if opts.Version[0] {
	// 	printVersion()
	// }

	switch opts.Loglevel[0] {
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "debug":
		level = log.DebugLevel
	default:
		level = log.DebugLevel
	}

	log.SetLevel(level)

	if len(opts.Influxhost) == 0 {
		log.Fatal("Missing influxhost parameter")
	}

	lastValues = make(map[string]float64)
	msgChannel = make(chan influxmessage, 30)

	mappingTable = readMappingFile(opts.Mappingfile[0])

	log.WithFields(log.Fields{
		"Influxhost":  opts.Influxhost[0],
		"Influxport":  opts.Influxport[0],
		"Influxdb":    opts.Influxdb[0],
		"Mappingfile": opts.Mappingfile[0],
		"loglevel":    level,
	}).Info("Starting up application")

	log.Info("Initiating handlers")

	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"requestURI":  r.RequestURI,
			"remoteAddr ": r.RemoteAddr,
		}).Error("failed to find page")
		// fmt.Println(r.RequestURI)
		http.ServeFile(w, r, "public/index.html")
	})

	r.HandleFunc("/temperature/{sensortype:[A-Za-z\\-]+}.{serial:[A-Z0-9]+}:{channel:[0-9]}.{type:[A-Z]+}/{value}", KlimaHandler)

	// r.HandleFunc("/kh/{serial:[A-Za-z0-9]+}.{type:[A-Za-z]+}/{value}", HistoryHandler).Methods("GET")
	r.HandleFunc("/status", Status)

	log.Info("starting channel sender")
	go channelSend()

	log.WithFields(log.Fields{
		"port": 4040,
		"host": "0.0.0.0",
	}).Info("starting listender")
	http.ListenAndServe("0.0.0.0:4040", r)

}

type status struct {
	running bool `json:"running"`
}

func Status(resp http.ResponseWriter, req *http.Request) {
	var retVal status
	retVal = status{running: true}

	statusString, err := json.Marshal(retVal)

	fmt.Println(statusString)

	if err != nil {
		panic(err)
	}

	fmt.Fprint(resp, statusString)
}

func KlimaHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	measurement := strings.ToLower(vars["type"])
	lowerSerial := strings.ToLower(vars["serial"])
	log.WithFields(log.Fields{
		"measurementtype": measurement,
		"serial":          lowerSerial,
		"channel":         vars["channel"],
		"sensortype":      vars["sensortype"],
	}).Debug("Received Measurement type")

	value, err := strconv.ParseFloat(vars["value"], 64)

	if err != nil {
		log.Fatal(err)
	}

	tags := make(map[string]string)

	customTags, ok := mappingTable[lowerSerial]

	if ok {
		for _, tag := range customTags {
			tagname, nameok := tag["name"]
			tagvalue, valueok := tag["value"]

			if nameok && valueok {
				tags[tagname] = tagvalue
				// fmt.Printf("%v: %v\n", tagname, tagvalue)
			}
		}
	}

	tags["serial"] = lowerSerial
	tags["channel"] = vars["channel"]
	tags["sensortype"] = vars["sensortype"]

	lastValues[lowerSerial] = value

	fields := map[string]interface{}{
		"value": value,
	}

	newInfluxMessage := newInfluxmessage(measurement, tags, fields)
	msgChannel <- newInfluxMessage

	resp.Header().Add("Content-Type", "text/html")
	resp.WriteHeader(http.StatusOK)
	fmt.Fprintf(resp, "OK! %v", time.Now())
	// go sendToInflux(measurement, vars["value"], tags)

}

func channelSend() {
	for {
		curr := <-msgChannel
		log.WithFields(log.Fields{
			"channelLength": len(msgChannel),
		}).Debug("Recieved message from channel")
		sendToInflux(curr)
	}
}

func sendToInflux(message influxmessage) {
	// var tryCount int
	// time.Sleep(time.Second * 3)

	fields := log.Fields{}

	for key, value := range message.tags {
		fields[key] = value
	}

	for key, value := range message.fields {
		fields[key] = value
	}

	fields["influxhost"] = opts.Influxhost[0]
	fields["influxport"] = opts.Influxport[0]
	fields["influxdb"] = opts.Influxdb[0]
	fields["measurement"] = message.measurement

	requestLogger := log.WithFields(fields)
	requestLogger.Debug("Metric recieved")

	hosturl := fmt.Sprintf("http://%v:%v", opts.Influxhost[0], opts.Influxport[0])

	// var count int = 2
	// var c client.client

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: hosturl,
	})
	defer c.Close()

	if err != nil {
		requestLogger.Error(err)
	} else {
		requestLogger.Debug("HTTP client created")
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  opts.Influxdb[0],
		Precision: "s",
	})

	if err != nil {
		requestLogger.Error(err)
	} else {
		requestLogger.Debug("BatchPoints created")
	}

	pt, err := client.NewPoint(message.measurement, message.tags, message.fields, message.time)
	if err != nil {
		fmt.Println("start sleeping 2")
		requestLogger.Error(err)
	} else {
		requestLogger.Debug("Point created")
	}
	bp.AddPoint(pt)

	var count int = 0
	for {
		if err := c.Write(bp); err != nil {
			count++
			// try to send a metric serval times with an waittime.
			if count < 8 {
				requestLogger.Warn(err)
				time.Sleep(time.Second * 1)
				continue
			} else {
				// If send fails continously the metric will be droped
				requestLogger.Errorf("Metric not send: %v", err)
			}
		} else {
			requestLogger.Info("Successfull send metric")
		}
		break
	}

}
