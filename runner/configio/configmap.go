package configio

import "time"

type Map map[string]interface{}

var Default = map[string]interface{}{
	"request": map[string]interface{}{
		"method": "GET",
		"url":    "http://localhost:8080/",
		"queryParams": map[string]interface{}{
			"foo": "bar",
		},
		"header": map[string]interface{}{
			"foo": "bar",
		},
		"body": map[string]interface{}{
			"type":    "raw",
			"content": "{}",
		},
	},
	"runner": map[string]interface{}{
		"requests":       1,
		"concurrency":    1,
		"interval":       "1s",
		"requestTimeout": "1s",
		"globalTimeout":  "1s",
	},
	"output": map[string]interface{}{
		"silent": false,
	},
	"tests": []map[string]interface{}{
		{
			"name":      "test1",
			"field":     "FAILURE_COUNT",
			"predicate": "EQ",
			"target":    0,
		},
		{
			"name":      "test2",
			"field":     "MAX",
			"predicate": "LT",
			"target":    200 * time.Millisecond,
		},
	},
}

func (m Map) Extra() map[string]interface{} {
	extra := map[string]interface{}{}
	for k, v := range m {
		if _, exists := Default[k]; !exists {
			extra[k] = v
		}
	}
	return extra
}
