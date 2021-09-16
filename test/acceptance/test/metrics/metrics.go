package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type record struct {
	Startdate       string `json:"startDate"`
	Enddate         string `json:"endDate"`
	IntegrationName string `json:"taskName"`
	Stage           string `json:"status"`
}

type Records struct {
	records []record
}

func NewRecords() *Records {
	return &Records{make([]record, 0)}
}

const JS_TIME_LAYOUT = "Mon Jan 02 15:04:05 MST 2006"

func (r *Records) AddRecords(start, end time.Time, integrationName, stage string) {
	st := start.Format(JS_TIME_LAYOUT)
	en := end.Format(JS_TIME_LAYOUT)
	r.records = append(r.records, record{fmt.Sprintf("_%s-", st), fmt.Sprintf(";%s+", en), integrationName, stage})
}

func (r *Records) GetJSArray() string {

	records := "["

	for _, rec := range r.records {
		bts, err := json.Marshal(rec)
		if err != nil {
			panic(err)
		}
		firstStart := strings.Index(string(bts), "_") - 2
		endStart := bytes.IndexByte(bts, '-')
		bts = append(bts[:firstStart+1], append([]byte(fmt.Sprintf("new Date(\"%s\")", rec.Startdate[1:len(rec.Startdate)-1])), bts[endStart+2:]...)...)

		firstStart = strings.Index(string(bts), ";") - 2
		endStart = bytes.IndexByte(bts, '+')
		bts = append(bts[:firstStart+1], append([]byte(fmt.Sprintf("new Date(\"%s\")", rec.Enddate[1:len(rec.Enddate)-1])), bts[endStart+2:]...)...)

		records += string(bts) + ",\n"

	}

	if len(r.records) != 0 {
		records = records[:len(records)-2]
	}

	records += "]"

	return records
}
