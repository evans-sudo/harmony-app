


package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const (
	redirectURI = "http://localhost:8080/callback"
	clientID    = "a9d17c4cfa3144b6aed7a78cd3fbfdae"
	clientSecret = "61cca6ef29b94614915f90c4a62432c0"
)

var (
	auth  = spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopePlaylistReadPrivate),
	)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	// Start the HTTP server for Spotify's OAuth callback
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Generate the authorization URL and prompt the user to log in
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// Wait for the Spotify client
	client := <-ch

	// Use the authenticated client
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}


// Feature 1: Show user details
func showUserDetails(client *spotify.Client) {
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user details: %v", err)
	}
	fmt.Printf("Logged in as: %s (ID: %s)\n", user.DisplayName, user.ID)
}

func createAndAddTracksToPlaylist(client *spotify.Client) {
	// Replace with actual track IDs
	trackIDs := []spotify.ID{
		"3n3Ppam7vgaVa1iaRUc9Lp", // Example track ID
		"7ouMYWpwJ422jRcDASZB7P", // Example track ID
	}

	// Create a new playlist
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch user: %v", err)
	}

	newPlaylist, err := client.CreatePlaylistForUser(
		context.Background(),
		user.ID,
		"My API Playlist",
		"A playlist created via Spotify API",
		false, // Public
		false, // Collaborative
	)
	if err != nil {
		log.Fatalf("Failed to create playlist: %v", err)
	}
	fmt.Printf("\nCreated Playlist: %s (ID: %s)\n", newPlaylist.Name, newPlaylist.ID)

	// Add tracks to the new playlist
	_, err = client.AddTracksToPlaylist(context.Background(), newPlaylist.ID, trackIDs...)
	if err != nil {
		log.Fatalf("Failed to add tracks to playlist: %v", err)
	}
	fmt.Println("Tracks added successfully!")
}


