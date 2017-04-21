package ccuprocessing

import (
	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
)

func QueryDB(clnt client.Client, query string, database string) (res []client.Result, err error) {

	log.WithFields(log.Fields{
		"function": "queryDB",
		"query":    query,
		"database": database,
	}).Debug("query about to execute")

	q := client.Query{
		Command:  query,
		Database: database,
	}

	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			log.WithFields(log.Fields{
				"function": "queryDB",
				"query":    query,
				"database": database,
			}).Errorf("Error in query response: %v", err)

			return res, response.Error()
		}
		res = response.Results
	} else {
		log.WithFields(log.Fields{
			"function": "queryDB",
			"query":    query,
			"database": database,
		}).Errorf("Error in query execution: %v", err)

		return res, err
	}

	log.WithFields(log.Fields{
		"function": "queryDB",
		"query":    query,
		"database": database,
	}).Debug("query successfull")

	return res, nil
}
