package services

import (
	"log"
	"os"

	"github.com/supabase-community/supabase-go"
)

var Client *supabase.Client

func InitSupabase() {
	client, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"), &supabase.ClientOptions{})
	if err != nil {
		log.Fatalf("Failed to create Supabase client: %v", err)
	} else {
		Client = client
	}
}
