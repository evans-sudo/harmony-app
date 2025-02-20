package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	auth   *spotifyauth.Authenticator
	ch     = make(chan *spotify.Client)
	state  = "abc123"
	client *spotify.Client
)

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redirectURI := os.Getenv("REDIRECT_URI")
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	auth = spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadPlaybackState,
		),
	)
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

	client = spotify.New(auth.Client(r.Context(), tok))
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "Login Completed!"+html)
	ch <- client
}

func searchTrack(client *spotify.Client, trackName string) ([]spotify.FullTrack, error) {
	results, err := client.Search(context.Background(), trackName, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}

	if len(results.Tracks.Tracks) > 0 {
		return results.Tracks.Tracks, nil
	}
	return nil, fmt.Errorf("no track found for: %s", trackName)
}

func playTrack(client *spotify.Client, trackURI spotify.URI) error {
	// List available devices
	devices, err := client.PlayerDevices(context.Background())
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return fmt.Errorf("no active devices found")
	}

	// Select the first available device
	deviceID := devices[0].ID

	err = client.PlayOpt(context.Background(), &spotify.PlayOptions{
		DeviceID: &deviceID,
		URIs:     []spotify.URI{trackURI},
	})
	if err != nil {
		return err
	}
	return nil
}

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
	if err != nil {
		log.Fatal(err)
	}

	for _, playlist := range playlists.Playlists {
		fmt.Println("Playlist name:", playlist.Name)
	}

	// Handle player actions
	http.HandleFunc("/player/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		action := strings.TrimPrefix(r.URL.Path, "/player/")
		fmt.Println("Got request for:", action)
		var err error
		switch action {
		case "play":
			trackURI := r.URL.Query().Get("uri")
			if trackURI != "" {
				err = playTrack(client, spotify.URI(trackURI))
			} else {
				err = client.Play(ctx)
			}
		case "pause":
			err = client.Pause(ctx)
		case "next":
			err = client.Next(ctx)
		case "previous":
			err = client.Previous(ctx)
		case "shuffle":
			playerState, err := client.PlayerState(ctx)
			if err != nil {
				log.Print(err)
			}
			_ = client.Shuffle(ctx, !playerState.ShuffleState)
		}
		if err != nil {
			log.Print(err)
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "Action completed")
	})

	// Handle search action
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		trackName := r.URL.Query().Get("track")
		if trackName == "" {
			http.Error(w, "Missing track parameter", http.StatusBadRequest)
			return
		}

		tracks, err := searchTrack(client, trackName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		for _, track := range tracks {
			fmt.Fprintf(w, "Track: %s by %s<br/>", track.Name, track.Artists[0].Name)
			fmt.Fprintf(w, "<a href=\"/player/play?uri=%s\">Play</a><br/><br/>", track.URI)
		}
	})

	// Keep the server running
	select {}
}

var html = `
<br/>
<a href="/player/play">Play</a><br/>
<a href="/player/pause">Pause</a><br/>
<a href="/player/next">Next track</a><br/>
<a href="/player/previous">Previous Track</a><br/>
<a href="/player/shuffle">Shuffle</a><br/>
<a href="/search?track=Shape%20of%20You">Search for ""</a><br/>
`
