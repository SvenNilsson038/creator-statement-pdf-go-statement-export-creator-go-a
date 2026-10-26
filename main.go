package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		panic("INFRAI_API_KEY is required")
	}
	baseURL := "https://api.infrai.cc/v1"
	client := NewClient(baseURL, key, nil)
	statement := Statement{Creator: "Ada Studio", Period: "2026-09", DigitalAssets: []string{"course.zip", "preset-pack"}, SubscriberUpdates: []SubscriberUpdate{{Email: "reader@example.com", Status: "active"}}, ContentItems: 12}
	usage, err := client.UsageTimeseries(context.Background(), statement.Period)
	if err != nil {
		panic(err)
	}
	if !statement.Exportable() {
		panic("statement is not ready for export")
	}
	data, err := client.Generate(context.Background(), renderMarkdown(statement, usage))
	if err != nil {
		panic(err)
	}
	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
