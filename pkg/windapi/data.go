package windapi

import (
	"encoding/json"
	"time"
)

// WindData is the output entry of wind's api
type WindData struct {
	UpdateTime time.Time
	WindCode   string

	Fields []string
	Values []interface{}

	CreatedAt time.Time
}

func (data *WindData) String() string {
	mp := make(map[string]interface{})
	for i := range data.Fields {
		mp[data.Fields[i]] = data.Values[i]
	}
	bts, err := json.Marshal(&struct {
		UpdateTime time.Time   `json:"UpdateTime"`
		WindCode   string      `json:"WindCode"`
		Data       interface{} `json:"Data"`
		CreatedAt  time.Time   `json:"Created"`
	}{
		UpdateTime: data.UpdateTime,
		WindCode:   data.WindCode,
		Data:       mp,
		CreatedAt:  data.CreatedAt,
	})
	if err != nil {
		return "wind: err, " + err.Error()
	}
	return string(bts)
}
