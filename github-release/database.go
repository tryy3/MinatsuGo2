package github_release

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/supabase-community/postgrest-go"
	realtimego "github.com/tryy3/realtime-go"
)

type EventRow struct {
	ID        int                    `json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	IsHandled bool                   `json:"is_handled"`
	RepoURL   string                 `json:"repo_url"`
	Event     string                 `json:"event"`
	RawData   map[string]interface{} `json:"raw_data"`
}

type Database struct {
	supabase_client   *postgrest.Client
	supabase_realtime *realtimego.Client
}

func (d *Database) GetEvent(id float64) (*EventRow, error) {
	data, _, err := d.supabase_client.
		From("events").
		Select("*", "1", false).
		Eq("id", strconv.Itoa(int(id))).
		Eq("is_handled", "false").
		Order("id", &postgrest.DefaultOrderOpts).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("error getting event from supabase: %v", err)
	}

	var event []EventRow
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling event: %v", err)
	}
	return &event[0], nil
}

func NewDatabase(m *AnnouncementManager) (*Database, error) {
	supabaseEndpoint := getSecretData("SupabaseEndpoint")
	if supabaseEndpoint == "" {
		return nil, fmt.Errorf("SupabaseEndpoint environment variable is empty")
	}

	supabaseAPIKey := getSecretData("SupabaseAPIKey")
	if supabaseAPIKey == "" {
		return nil, fmt.Errorf("SupabaseAPIKey environment variable is empty")
	}

	client := postgrest.NewClient(fmt.Sprintf("https://%s/rest/v1", supabaseEndpoint), "", map[string]string{
		"apikey":        supabaseAPIKey,
		"Authorization": fmt.Sprintf("Bearer %s", supabaseAPIKey),
	})
	if client.ClientError != nil {
		return nil, fmt.Errorf("error creating supabase client: %v", client.ClientError)
	}

	// create client
	c, err := realtimego.NewClient(supabaseEndpoint, supabaseAPIKey) // realtimego.WithUserToken(RLS_TOKEN),

	if err != nil {
		return nil, fmt.Errorf("error creating realtime client: %v", err)
	}

	// connect to server
	err = c.Connect()
	if err != nil {
		return nil, fmt.Errorf("error connecting to realtime server: %v", err)
	}

	// create and subscribe to channel
	db := "realtime"
	schema := "public"
	table := "events"
	ch, err := c.Channel(realtimego.WithTable(&db, &schema, &table))
	if err != nil {
		return nil, fmt.Errorf("error creating channel: %v", err)
	}

	// setup hooks
	ch.OnDelete = func(m realtimego.Message) {
		log.Println("***ON DELETE....", m)
	}

	ch.OnUpdate = m.DatabaseUpdate
	ch.OnInsert = m.DatabaseUpdate

	// subscribe to channel
	err = ch.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("error subscribing to channel: %v", err)
	}
	return &Database{supabase_realtime: c, supabase_client: client}, nil
}
