package routes

import (
	"database/sql"
	"github.com/Fyefhqdishka/song-library/controllers"
	"github.com/Fyefhqdishka/song-library/repositories"
	"github.com/gorilla/mux"
	"log/slog"
)

func RegisterRoutes(r *mux.Router, db *sql.DB, logger *slog.Logger) {
	songRepo := repositories.NewSongRepository(db, logger)
	songController := controllers.NewSongController(songRepo, logger)
	r.HandleFunc("/songs", songController.AddSongs).Methods("POST")
	r.HandleFunc("/songs", songController.DeleteSongs).Methods("DELETE")
	r.HandleFunc("/songs", songController.GetSongs).Methods("GET")
	r.HandleFunc("/songs/verses", songController.GetSongsVerses).Methods("GET")
	r.HandleFunc("/update-text", songController.UpdateSongsText).Methods("PUT")
}
