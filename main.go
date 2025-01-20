package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const (
	redirectURI  = "http://localhost:8080/callback"
	clientID     = "a9d17c4cfa3144b6aed7a78cd3fbfdae"
	clientSecret = "61cca6ef29b94614915f90c4a62432c0"
)

var (
	auth = spotifyauth.New(
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

	// Check for granted scopes
	scopes := os.Getenv("SPOTIFY_SCOPES")
	if scopes == "" {
		fmt.Println("No scopes found. Ensure the necessary permissions are granted.")
	} else {
		fmt.Println("Granted Scopes:", scopes)
	}

	// Wait for the Spotify client
	client := <-ch

	// Use the authenticated client
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID, user.DisplayName)


	    // Retrieve and display user's playlists
		playlists, err := client.CurrentUsersPlaylists(context.Background())
		if (err != nil) {
			log.Fatal(err)
		}
	
		for _, playlist := range playlists.Playlists {
			fmt.Println("Playlist name:", playlist.Name)
		}

		// Search and play a track
		searchAndPlayTrack(client, "Shape of You")
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


func searchAndPlayTrack(client *spotify.Client, trackName string) {
	results, err := client.Search(context.Background(), trackName, spotify.SearchTypeTrack)
    if err != nil {
        log.Fatal(err)
    }

    if len(results.Tracks.Tracks) > 0 {
        track := results.Tracks.Tracks[0]
        fmt.Println("Playing track:", track.Name, "by", track.Artists[0].Name)
		err = client.PlayOpt(context.Background(), &spotify.PlayOptions{
			URIs: []spotify.URI{track.URI},
		})
        if err != nil {
            log.Fatal(err)
        }
    } else {
        fmt.Println("No track found for:", trackName)
    }
}
