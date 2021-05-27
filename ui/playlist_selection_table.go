package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/firstlane/baton/api"
	"github.com/firstlane/baton/utils"
	"github.com/jroimartin/gocui"
)

// PlaylistSelectionTable implements the Table interface for "Simple" Playlist objects as defined by the Spotify Web API
type PlaylistSelectionTable struct {
	playlists  *api.SimplePlaylistsPaged
	selections []bool
}

// NewPlaylistSelectionTable creates a new instance of PlaylistSelectionTable
func NewPlaylistSelectionTable(playlistsPaged *api.SimplePlaylistsPaged) *PlaylistSelectionTable {
	return &PlaylistSelectionTable{
		playlists:  playlistsPaged,
		selections: make([]bool, len(playlistsPaged.Items)),
	}
}

func (p *PlaylistSelectionTable) getColumnWidths(maxX int) map[string]int {
	m := make(map[string]int)
	m["name"] = maxX / 3
	m["owner"] = maxX / 6
	m["total"] = maxX / 8
	m["selected"] = maxX - m["name"] - m["owner"] - m["total"]

	return m
}

func (p *PlaylistSelectionTable) renderHeader(v *gocui.View, maxX int) {
	columnWidths := p.getColumnWidths(maxX)

	selectedHeader := utils.LeftPaddedString("SELECTED", columnWidths["selected"], 2)
	nameHeader := utils.LeftPaddedString("NAME", columnWidths["name"], 2)
	totalHeader := utils.LeftPaddedString("TOTAL", columnWidths["total"], 2)
	ownerHeader := utils.LeftPaddedString("OWNER", columnWidths["owner"], 2)

	fmt.Fprintf(v, "\u001b[1m%s[0m\n", utils.LeftPaddedString("PLAYLISTS", maxX, 2))
	fmt.Fprintf(v, "\u001b[1m%s %s %s %s\u001b[0m\n", selectedHeader, nameHeader, ownerHeader, totalHeader)
}

func (p *PlaylistSelectionTable) render(v *gocui.View, maxX int) {
	columnWidths := p.getColumnWidths(maxX)

	for index, playlist := range p.playlists.Items {
		selectedText := "False"

		if p.selections[index] {
			selectedText = "True"
		}

		selected := utils.LeftPaddedString(selectedText, columnWidths["selected"], 2)
		name := utils.LeftPaddedString(playlist.Name, columnWidths["name"], 2)
		owner := utils.LeftPaddedString(playlist.Owner.DisplayName, columnWidths["owner"], 2)
		total := utils.LeftPaddedString(strconv.Itoa(playlist.Tracks.Total), columnWidths["total"], 2)

		fmt.Fprintf(v, "\n%s %s %s %s", selected, name, owner, total)
	}
}

func (p *PlaylistSelectionTable) renderFooter(v *gocui.View, maxX int) {
	fmt.Fprintf(v, "\u001b[1m%s\u001b[0m\n", utils.LeftPaddedString(fmt.Sprintf("Showing %d of %d playlists", len(p.playlists.Items), p.playlists.Total), maxX, 2))
}

func (p *PlaylistSelectionTable) getTableLength() int {
	return len(p.playlists.Items)
}

func (p *PlaylistSelectionTable) loadNextRecords() error {
	if p.playlists.Next != "" {
		if strings.Contains(p.playlists.Next, "api.spotify.com/v1/search") {
			res, err := api.GetNextSearchResults(p.playlists.Next)

			if err != nil {
				return err
			}

			nextPlaylists := res.Playlists

			p.playlists.Href = nextPlaylists.Href
			p.playlists.Offset = nextPlaylists.Offset
			p.playlists.Next = nextPlaylists.Next
			p.playlists.Previous = nextPlaylists.Previous
			p.playlists.Items = append(p.playlists.Items, nextPlaylists.Items...)
		} else {
			res, err := api.GetNextMyPlaylists(p.playlists.Next)

			if err != nil {
				return err
			}

			nextPlaylists := res

			p.playlists.Href = nextPlaylists.Href
			p.playlists.Offset = nextPlaylists.Offset
			p.playlists.Next = nextPlaylists.Next
			p.playlists.Previous = nextPlaylists.Previous
			p.playlists.Items = append(p.playlists.Items, nextPlaylists.Items...)
		}

	}

	return nil
}

func (p *PlaylistSelectionTable) playSelected(selectedIndex int) (string, error) {
	playlist := p.playlists.Items[selectedIndex]
	playerOptions := api.PlayerOptions{
		ContextURI: playlist.URI,
	}

	chosenItem := fmt.Sprintf("Now playing the playlist: %s by %s\n", playlist.Name, playlist.Owner.DisplayName)

	return chosenItem, api.StartPlayback(&playerOptions)
}

func (p *PlaylistSelectionTable) newTableFromSelection(selectedIndex int) (Table, error) {
	playlist := p.playlists.Items[selectedIndex]
	tracksPaged, err := api.GetTracksForPlaylist(playlist.Owner.ID, playlist.ID)

	if err != nil {
		return nil, err
	}

	return NewPlaylistTrackTable(&tracksPaged, &playlist), nil
}

func (p *PlaylistSelectionTable) handleSaveKey(selectedIndex int) error {
	if p.selections[selectedIndex] {
		// Toggle off
		p.selections[selectedIndex] = false
	} else {
		// Toggle on
		p.selections[selectedIndex] = true
	}

	// Save current selections to file

	return nil
}
