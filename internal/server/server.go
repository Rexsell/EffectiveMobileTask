package server

import (
	"EffectiveMobileTask/internal/database"
	"net/http"
)

type Server struct {
	Port string
}

func New(port string, db *database.DB) *Server {
	DB = db
	return &Server{
		Port: port,
	}
}

func (s *Server) StartServer() error {
	http.HandleFunc("/info", SongInfoHandler)

	http.HandleFunc("/fullinfo", SongFullInfoHandler)

	http.HandleFunc("/getInfoByField", SongsByFieldHandler)

	http.HandleFunc("/getText", SongTextHandler)

	http.HandleFunc("/delete", DeleteSongHandler)

	http.HandleFunc("/edit", EditSongHandler)

	http.HandleFunc("/add", AddSongHandler)
	err := http.ListenAndServe(s.Port, nil)
	if err != nil {
		return err
	}
	return nil
}
