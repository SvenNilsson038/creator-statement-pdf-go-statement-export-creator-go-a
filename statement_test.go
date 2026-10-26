package main

import "testing"

func TestStatementExportableRequiresSubscriberUpdate(t *testing.T) {
	base := Statement{Creator: "Ada Studio", Period: "2026-09", DigitalAssets: []string{"course.zip"}, ContentItems: 1}
	tests := []struct {
		name    string
		updates []SubscriberUpdate
		want    bool
	}{
		{"ready", []SubscriberUpdate{{Email: "reader@example.com", Status: "active"}}, true},
		{"pending subscriber sync", nil, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base.SubscriberUpdates = tc.updates
			if got := base.Exportable(); got != tc.want {
				t.Fatalf("Exportable() = %v, want %v", got, tc.want)
			}
		})
	}
}
