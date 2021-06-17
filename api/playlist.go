package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/firstlane/baton/utils"
	"github.com/google/go-querystring/query"
)

// The PlaylistTrackLinks struct describes a Playlist Track Link object as defined by the Spotify Web API
type PlaylistTrackLinks struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

// The SimplePlaylist struct describes a "Simple" Playlist object as defined by the Spotify Web API
type SimplePlaylist struct {
	Collaborative bool                `json:"collaborative"`
	ExternalUrls  map[string]string   `json:"external_urls"`
	Href          string              `json:"href"`
	ID            string              `json:"id"`
	Images        []Image             `json:"images"`
	Name          string              `json:"name"`
	Owner         *User               `json:"owner"`
	Public        bool                `json:"public"`
	SnapshotID    string              `json:"snapshot_id"`
	Tracks        *PlaylistTrackLinks `json:"tracks"`
	Type          string              `json:"type"`
	URI           string              `json:"uri"`
}

// The SimplePlaylistsPaged struct is a slice of SimplePlaylist objects wrapped in a Spotify paging object
type SimplePlaylistsPaged struct {
	Href     string           `json:"href"`
	Items    []SimplePlaylist `json:"items"`
	Limit    int              `json:"limit"`
	Next     string           `json:"next"`
	Offset   int              `json:"offset"`
	Previous string           `json:"previous"`
	Total    int              `json:"total"`
}

// GetTracksForPlaylist returns a list of PlaylistTrack objects in a paging object for the given user and playlist
func GetTracksForPlaylist(userID, playlistID string) (pt PlaylistTracksPaged, err error) {
	t := getAccessToken()

	r := buildRequest("GET", apiURLBase+"users/"+userID+"/playlists/"+playlistID+"/tracks", nil, nil)
	r.Header.Add("Authorization", "Bearer "+t)

	err = makeRequest(r, &pt)

	return pt, err
}

// GetAllTracksForPlaylist returns a list of PlaylistTrack objects in a paging object for the given user and playlist
func GetAllTracksForPlaylist(userID, playlistID string) (pt PlaylistTracksPaged, err error) {
	playlistTracks, err := GetTracksForPlaylist(userID, playlistID)

	if err != nil {
		fmt.Println("Failed to get all tracks for playlist ", playlistID)
		return
	}

	for {
		if err != nil || playlistTracks.Offset >= playlistTracks.Total-1 {
			break
		}

		nextTracks, err := GetNextTracksForPlaylist(playlistTracks.Next)

		if err != nil {
			break
		}

		playlistTracks.Href = nextTracks.Href
		playlistTracks.Offset = nextTracks.Offset
		playlistTracks.Next = nextTracks.Next
		playlistTracks.Previous = nextTracks.Previous
		playlistTracks.Items = append(playlistTracks.Items, nextTracks.Items...)
	}

	pt = playlistTracks
	return pt, err
}

// GetNextTracksForPlaylist takes in the Next field from the paging objects returned from GetTracksForPlaylist and allows you to move forward through the tracks
func GetNextTracksForPlaylist(url string) (pt PlaylistTracksPaged, err error) {
	t := getAccessToken()

	r, err := http.NewRequest("GET", url, nil)
	r.Header.Add("Authorization", "Bearer "+t)

	err = makeRequest(r, &pt)

	return pt, err
}

// GetMyPlaylists takes in the Next field from the paging objects returned from GetTracksForPlaylist and allows you to move forward through the tracks
func GetMyPlaylists() (pt *SimplePlaylistsPaged, err error) {
	v, err := query.Values(nil)

	if err != nil {
		return pt, err
	}

	// These are the defaults but are required, otherwise spotify will not return displaynames for owners
	v.Set("limit", "10")
	v.Set("offset", "0")

	t := getAccessToken()

	r := buildRequest("GET", apiURLBase+"me/playlists", v, nil)
	r.Header.Add("Authorization", "Bearer "+t)

	err = makeRequest(r, &pt)

	return pt, err
}

// GetNextMyPlaylists takes in the Next fields from the paging objects returned from me/playlists and allows you to move forward through the results
func GetNextMyPlaylists(url string) (pt *SimplePlaylistsPaged, err error) {
	t := getAccessToken()

	r, err := http.NewRequest("GET", url, nil)
	r.Header.Add("Authorization", "Bearer "+t)

	err = makeRequest(r, &pt)

	return pt, err
}

func loadNextRecords(playlists *SimplePlaylistsPaged) error {
	if playlists.Next != "" {
		if strings.Contains(playlists.Next, "api.spotify.com/v1/search") {
			return errors.New("I dunno, something happened")
		}

		res, err := GetNextMyPlaylists(playlists.Next)

		if err != nil {
			return err
		}

		nextPlaylists := res

		playlists.Href = nextPlaylists.Href
		playlists.Offset = nextPlaylists.Offset
		playlists.Next = nextPlaylists.Next
		playlists.Previous = nextPlaylists.Previous
		playlists.Items = append(playlists.Items, nextPlaylists.Items...)
	}

	return nil
}

// GetAllMyPlaylists takes in the Next field from the paging objects returned from GetTracksForPlaylist and allows you to move forward through the tracks
func GetAllMyPlaylists() (playlists *SimplePlaylistsPaged, err error) {
	quit := make(chan bool)

	go utils.StartProgressSpinner(quit)

	playlists, err = GetMyPlaylists()

	const limit = 10
	playlistCounter := limit // This is the default limit set in GetMyPlaylists when calling the Web API

	if err == nil && playlists.Total > playlistCounter {
		for {
			err = loadNextRecords(playlists)
			playlistCounter += limit
			if err != nil || playlistCounter >= playlists.Total {
				break
			}
		}
	} else {
		fmt.Println("Did not enter for loop in GetAllMyPlaylists")
		fmt.Println("err = ", err)
	}

	quit <- true

	return playlists, err
}
