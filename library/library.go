package library

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/firstlane/baton/api"
	"github.com/firstlane/baton/ui"
	"github.com/spf13/cobra"
)

/*
 * TODO: Need functions for the following:
 *	- Loading the json library file
 *	- Writing to the json library file
 *	- Accessors for info from the json data
 *	- Get all songs that are not recorded
 *
 * TODO: Need following functionality:
 *	- Do not remove song from playlist if it is no longer in the Spotify copy of the playlist.
 *	- Add extra fields for:
 *		* Recorded?
 *		* Recording date
 *		* Recording quality
 */

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getPlaylistsCmd)
	//getPlaylistsCmd.AddCommand(updatePlaylistsCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download JSON data from your library",
	Long:  `Download JSON data from your library`,
	Run:   getLibrary,
}

func getTracksFromPlaylist(playlist api.SimplePlaylist, tracks map[string]api.FullTrack) {
	playlistTracks, err := api.GetAllTracksForPlaylist(playlist.Owner.ID, playlist.ID)

	if err != nil {
		fmt.Println("Failed to get tracks from playlist ", playlist.Name)
		return
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

		_, exists := tracks[track.Track.ID]
		if !exists {
			var newTrack = track.Track

			var newPlaylists = append(newTrack.Playlists, trackPlaylist)
			newTrack.Playlists = newPlaylists

			tracks[track.Track.ID] = newTrack
		} else {
			// TODO: What if a playlist has multiple instances of the same song?
			var newTrack = tracks[track.Track.ID]

			var newPlaylists = append(newTrack.Playlists, trackPlaylist)
			newTrack.Playlists = newPlaylists

			tracks[track.Track.ID] = newTrack
		}
	}
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
	fmt.Println("Getting each song from playlists")
	for _, playlist := range playlists.Items {
		fmt.Println("playlist = ", playlist.Name)

		getTracksFromPlaylist(playlist, allTracks)
	}

	// TODO: Load old json data if it exists and update accordingly

	jsonData, err := json.MarshalIndent(allTracks, "", "    ")
	if err != nil {
		fmt.Println("err", err)
	}

	_ = ioutil.WriteFile("test.json", jsonData, 0644)
}

func getPlayLists(cmd *cobra.Command, args []string) {

	res, err := api.GetAllMyPlaylists()

	if err != nil {

		fmt.Printf("Couldn't get your playlists from spotify. Have you authenticated with the 'auth' command?\n")
		fmt.Println("err", res, err)
		return
	}

	at := ui.NewPlaylistSelectionTable(res)

	err = ui.Run(at)

	if err != nil {
		log.Fatal(err)
	}

	var allTracks = make(map[string]api.FullTrack)

	for index, selection := range at.Selections {
		if selection {
			fmt.Println(at.Playlists.Items[index].Name, "is selected")
			getTracksFromPlaylist(at.Playlists.Items[index], allTracks)
		}
	}

	jsonData, err := json.MarshalIndent(allTracks, "", "    ")
	if err != nil {
		fmt.Println("err", err)
	}

	_ = ioutil.WriteFile("small-test.json", jsonData, 0644)
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
