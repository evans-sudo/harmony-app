# Harmony App

Harmony App is a music application that leverages Spotify's API to provide users with an enhanced music experience.

## Features

- Search for songs, albums, and artists
- Create and manage playlists
- Discover new music based on your preferences
- Stream music directly from Spotify

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/sudo-evans/harmony-app.git
    ```
2. Navigate to the project directory:
    ```bash
    cd harmony-app
    ```
3. Install dependencies:
    ```bash
    go mod init
    ```

## Usage

1. Obtain your Spotify API credentials by creating an app on the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications).
2. Create a `.env` file in the root directory and add your credentials:
    ```env
    SPOTIFY_CLIENT_ID=your_client_id
    SPOTIFY_CLIENT_SECRET=your_client_secret
    ```
3. Start the application:
    ```bash
    go run main.go
    ```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.