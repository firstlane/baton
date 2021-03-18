package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/firstlane/baton/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
	//getCmd.AddCommand(getPlaylistsCmd)
	//getPlaylistsCmd.AddCommand(updatePlaylistsCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download JSON data from your library",
	Long:  `Download JSON data from your library`,
	Run:   getLibrary,
}

func getLibrary(cmd *cobra.Command, args []string) {
	playlists, err := api.GetAllMyPlaylists()

	if err != nil {

		fmt.Printf("Couldn't get your playlists from spotify. Have you authenticated with the 'auth' command?\n")
		fmt.Println("err", err)
		return
	}

	// This is a mapping from Track ID to the FullTrack data
	var allTracks = make(map[string]api.FullTrack)

	// Get each song from playlists
	for _, playlist := range playlists.Items {
		playlistTracks, err := api.GetAllTracksForPlaylist(playlist.Owner.ID, playlist.ID)

		if err != nil {
			break
		}

		for _, track := range playlistTracks.Items {
			var trackPlaylist api.TrackPlaylistData
			trackPlaylist.AddedAt = track.AddedAt
			trackPlaylist.AddedBy = track.AddedBy
			trackPlaylist.Href = playlist.Href
			trackPlaylist.ID = playlist.ID
			trackPlaylist.IsLocal = track.IsLocal
			trackPlaylist.Name = playlist.Name
			trackPlaylist.Owner = playlist.Owner
			trackPlaylist.URI = playlist.URI

			_, exists := allTracks[track.Track.ID]
			if !exists {
				var newTrack = track.Track

				var newPlaylists = append(newTrack.Playlists, trackPlaylist)
				newTrack.Playlists = newPlaylists

				allTracks[track.Track.ID] = newTrack
			} else {
				// TODO: What if a playlist has multiple instances of the same song?
				var newTrack = allTracks[track.Track.ID]

				var newPlaylists = append(newTrack.Playlists, trackPlaylist)
				newTrack.Playlists = newPlaylists

				allTracks[track.Track.ID] = newTrack
			}
		}
	}

	// Load old json data if it exists and update accordingly
}

func getPlayLists(cmd *cobra.Command, args []string) {
	playlists, err := api.GetAllMyPlaylists()

	if err != nil {

		fmt.Printf("Couldn't get your playlists from spotify. Have you authenticated with the 'auth' command?\n")
		fmt.Println("err", err)
		return
	}

	jsonData, err := json.MarshalIndent(playlists, "", "\t")

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("test.json", jsonData, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func updatePlayLists(cmd *cobra.Command, args []string) {

	// Check for local JSON data. Err if no data.

	// if err != nil {
	// 	fmt.Printf("Could not find any downloaded playlist data to update. Try the 'get playlists' command to download some data.\n")
	// 	fmt.Println("err", res, err)
	// 	return
	// }

	res, err := api.GetMyPlaylists()

	if err != nil {

		fmt.Printf("Couldn't get your playlists from spotify. Have you authenticated with the 'auth' command?\n")
		fmt.Println("err", res, err)
		return
	}

	// at := ui.NewPlaylistTable(res)

	// err = ui.Run(at)

	if err != nil {
		log.Fatal(err)
	}
}

var getPlaylistsCmd = &cobra.Command{
	Use:   `playlists`,
	Short: "Choose playlists to download data from",
	Long:  `Choose playlists to download data from`,
	Run:   getPlayLists,
}

var updatePlaylistsCmd = &cobra.Command{
	Use:   `update`,
	Short: "Updates data for playlists that have been previously downloaded",
	Long:  `Updates data for playlists that have been previously downloaded`,
	Run:   updatePlayLists,
}
