package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func SongInfoHandler(w http.ResponseWriter, r *http.Request) {
	groupName, songName := parseParams(r)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	//	execute
	song, err := songInfo(ctx, songName, groupName)
	if err != nil {
		log.Errorf("obtain song info err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Printf("parsed song: %v", song)
	songToSend := &SongToSend{
		ReleaseDate: song.ReleaseDate,
		Text:        song.Text,
		Link:        song.Link,
	}
	//	send json
	if err := sendJson(w, http.StatusOK, songToSend); err != nil {
		log.Fatal(err)
	}
}

func SongFullInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	songId := params.Get("id")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	//	execute
	song, err := getSong(ctx, songId)
	if err != nil {
		log.Errorf("song full info obtaining err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Printf("parsed song: %v", song)
	//	send json
	if err := sendJson(w, http.StatusOK, song); err != nil {
		log.Fatal(err)
	}
}

func SongsByFieldHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	filterField := params.Get("field")
	log.Printf("parsed field: %s", filterField)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	songs, err := getSongsByField(ctx, filterField)
	if err != nil {
		log.Errorf("song info sort by field err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := sendJson(w, http.StatusOK, songs); err != nil {
		log.Fatal(err)
	}
}

func SongTextHandler(w http.ResponseWriter, r *http.Request) {
	// 	read data

	params := r.URL.Query()
	songId := params.Get("id")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	song, err := getSong(ctx, songId)
	if err != nil {
		log.Errorf("getting song text err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("parsed song: %v", song)

	verses := songTextVerses(song.Text)

	if err := sendJson(w, http.StatusOK, verses); err != nil {
		log.Fatal(err)
	}
}

func DeleteSongHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	songId := params.Get("id")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	song, err := getSong(ctx, songId)
	if err != nil {
		log.Errorf("getting song err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}

	err = deleteSong(ctx, songId)
	if err != nil {
		log.Errorf("deleting song err : %s", err.Error())
		if err := sendJson(w, http.StatusInternalServerError, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}

	if err := sendJson(w, http.StatusOK, "song deleted"); err != nil {
		log.Fatal(err)
	}
	log.Printf("song deleted: %s , group: %s", song.Title, song.GroupName)
}

func EditSongHandler(w http.ResponseWriter, r *http.Request) {
	// 	read data
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	var song *Song
	err = json.Unmarshal(body, &song)
	if err != nil {
		log.Errorf("unmarshal song in edit func err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = editSong(ctx, song)
	if err != nil {
		log.Errorf("edit song err : %s", err.Error())
		if err := sendJson(w, http.StatusInternalServerError, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}
	if err := sendJson(w, http.StatusOK, "song edited"); err != nil {
		log.Fatal(err)
	}
	log.Printf("song edited: %v", song)
}

func AddSongHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	var song *Song
	err = json.Unmarshal(body, &song)
	if err != nil {
		log.Errorf("unmarshalling to add song err : %s", err.Error())
		if err := sendJson(w, http.StatusBadRequest, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = addSong(ctx, song)
	if err != nil {
		log.Errorf("adding song err : %s", err.Error())
		if err := sendJson(w, http.StatusInternalServerError, ErrResponse{Error: err.Error()}); err != nil {
			log.Fatal(err)
		}
	}
	if err := sendJson(w, http.StatusOK, song); err != nil {
		log.Fatal(err)
	}
	log.Printf("song added: %v", song)
}

func getSong(ctx context.Context, id string) (*Song, error) {

	query := fmt.Sprintf("SELECT * FROM songs WHERE id = %s", id)

	row := DB.Conn.QueryRow(ctx, query)

	song := &Song{}
	err := row.Scan(&song.ID, &song.ReleaseDate, &song.Text, &song.Link, &song.Title, &song.GroupName)
	song.Text = strings.Replace(song.Text, "\\\\n", "\n", -1)
	if err != nil {
		log.Errorf("select song err: %s", err)
		return nil, err
	}
	return song, nil
}

func songInfo(ctx context.Context, songName string, groupName string) (*Song, error) {
	query := fmt.Sprintf("SELECT * FROM songs WHERE title = '%s' AND group_name = '%s'", songName, groupName)
	row := DB.Conn.QueryRow(ctx, query)

	song := &Song{}
	err := row.Scan(&song.ID, &song.ReleaseDate, &song.Text, &song.Link, &song.Title, &song.GroupName)
	if err != nil {
		log.Errorf("select song err: %s", err)
		return nil, err
	}
	song.Text = textFormat(song.Text)
	return song, nil
}

func getSongsByField(ctx context.Context, fieldToFilter string) ([]*Song, error) {

	query := fmt.Sprintf("SELECT * FROM songs ORDER BY %v", fieldToFilter)

	rows, err := DB.Conn.Query(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Errorf("select songs err: %s", err)
		return nil, err
	}

	songs := make([]*Song, 0)

	for rows.Next() {
		song := &Song{}
		err := rows.Scan(&song.ID, &song.ReleaseDate, &song.Text, &song.Link, &song.Title, &song.GroupName)
		if err != nil {
			log.Printf("parse songs fields err: %s", err)
			return nil, err
		}
		song.Text = strings.Replace(song.Text, "\\\\n", "\n", -1)
		log.Println(song)
		songs = append(songs, song)
	}

	return songs, nil
}

func deleteSong(ctx context.Context, songId string) error {

	query := fmt.Sprintf("DELETE FROM songs WHERE id = '%s'", songId)

	_, err := DB.Conn.Exec(ctx, query)
	return err
}

func editSong(ctx context.Context, song *Song) error {
	query := fmt.Sprintf("SELECT id FROM songs WHERE id='%d'", song.ID)
	row := DB.Conn.QueryRow(ctx, query)

	err := row.Scan()
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}

	query = "UPDATE songs SET release_date = $1, text = $2, link = $3, title = $4, group_name = $5 WHERE id = $6"
	_, err = DB.Conn.Exec(ctx,
		query,
		song.ReleaseDate,
		song.Text,
		song.Link,
		song.Title,
		song.GroupName,
		song.ID)
	return err
}

func addSong(ctx context.Context, song *Song) error {
	query := "INSERT INTO songs (release_date, group_name, text, link, title) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := DB.Conn.QueryRow(ctx, query, song.ReleaseDate, song.GroupName, song.Text, song.Link, song.Title).Scan(&song.ID)
	if err != nil {
		log.Errorf("adding song err : %s", err.Error())
		return err
	}
	return nil
}

func songTextVerses(songText string) []*Verse {
	songText = textFormat(songText)
	versesSplit := strings.Split(songText, "\n\n")
	verses := make([]*Verse, 0)
	for idx, verse := range versesSplit {
		verses = append(verses, &Verse{Text: verse, ID: idx + 1})
	}
	return verses
}

func textFormat(text string) string {
	return strings.Replace(text, "\\\\n", "\n", -1)
}

func sendJson(w http.ResponseWriter, status int, r any) error {
	b, err := json.Marshal(r)
	if err != nil {
		log.Errorf("marshal data err : %s", err.Error())
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(b)
	if err != nil {
		log.Errorf("send data err : %s", err.Error())
	}
	return err
}

func parseParams(r *http.Request) (string, string) {

	params := r.URL.Query()
	groupName := params.Get("group")
	songName := params.Get("song")
	log.Printf("group name: %s, song name: %s", groupName, songName)
	return groupName, songName
}
