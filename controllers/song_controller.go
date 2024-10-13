package controllers

import (
	"encoding/json"
	"github.com/Fyefhqdishka/song-library/models"
	"github.com/Fyefhqdishka/song-library/repositories"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SongController struct {
	repo   repositories.SongMethods
	Logger *slog.Logger
}

func NewSongController(repo *repositories.SongRepository, logger *slog.Logger) *SongController {
	return &SongController{repo: repo, Logger: logger}
}

func (m *SongController) AddSongs(w http.ResponseWriter, r *http.Request) {
	m.Logger.Info("Начало обработки запроса на добавление песни")

	var song models.Song

	releaseDateStr := r.URL.Query().Get("release_date")

	if releaseDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", releaseDateStr) // парсим в удобный формат
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			m.Logger.Error("ошибка парсинга даты: AddSongs", "error:", err)
			return
		}
		song.ReleaseDate = parsedDate
	}

	err := json.NewDecoder(r.Body).Decode(&song)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		m.Logger.Error("ошибка декодирования json", "error:", err)
		return
	}

	if song.Text != "" {
		song.Verses = strings.Split(song.Text, "\n\n") // разбиваем на куплеты
	}

	m.Logger.Info("Добавление песни в бд")
	err = m.repo.AddSong(&song)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		m.Logger.Error("Ошибка при добавлении песни в БД", "error:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(song)
}

func (m *SongController) DeleteSongs(w http.ResponseWriter, r *http.Request) {
	m.Logger.Info("Обработка запроса на удаление песни")

	var song models.Song
	err := json.NewDecoder(r.Body).Decode(&song)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		m.Logger.Error("Ошибка декодирования json", "error:", err)
		return
	}

	err = m.repo.DeleteSong(song.GroupName, song.SongTitle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		m.Logger.Error("Ошибка удаления песни из БД", "error:", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (m *SongController) GetSongs(w http.ResponseWriter, r *http.Request) {
	m.Logger.Info("начало обработки запроса на получение песен")

	groupName := r.URL.Query().Get("group_name")
	songTitle := r.URL.Query().Get("song_title")
	releaseDateStr := r.URL.Query().Get("release_date")
	text := r.URL.Query().Get("text")
	link := r.URL.Query().Get("link")

	var releaseDate time.Time
	if releaseDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", releaseDateStr) // парсим в удобный формат
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			m.Logger.Error("ошибка парсинга даты: GetSongs", "error:", err)
			return
		}
		releaseDate = parsedDate
	}

	limit := 10
	offset := 0

	songs, err := m.repo.GetSong(&models.GetSong{
		GroupName:   groupName,
		SongTitle:   songTitle,
		ReleaseDate: releaseDate,
		Text:        text,
		Link:        link,
	}, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		m.Logger.Error("ошибка получения песни из БД", "error:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(songs)
}

func (m *SongController) GetSongsVerses(w http.ResponseWriter, r *http.Request) {
	m.Logger.Debug("Начало получения куплетов")

	songTitle := r.URL.Query().Get("song_title")

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
		m.Logger.Debug("Используется значение по умолчанию для limit", "limit", limit)
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
		m.Logger.Debug("Используется значение по умолчанию для offset", "offset", offset)
	}

	verses, err := m.repo.GetSongVerses(songTitle, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		m.Logger.Error("Ошибка получения куплетов из БД", "error", err)
		return
	}

	m.Logger.Debug("Полученные куплеты:", "количество куплетов", len(verses))

	response := map[string]interface{}{
		"verses": verses,
		"count":  len(verses),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *SongController) UpdateSongsText(w http.ResponseWriter, r *http.Request) {
	m.Logger.Debug("Начало обновления текста песни")

	songTitle := r.URL.Query().Get("song_title")
	if songTitle == "" {
		http.Error(w, "song_title обязателен", http.StatusBadRequest)
		return
	}

	var requestData struct {
		NewText string `json:"new_text"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		m.Logger.Error("Ошибка декодирования данных", "error", err)
		return
	}

	err = m.repo.UpdateSongText(songTitle, requestData.NewText)
	if err != nil {
		http.Error(w, "Ошибка при обновлении текста песни", http.StatusInternalServerError)
		m.Logger.Error("Ошибка обновления текста песни в БД", "error", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Текст успешно обновлен"})
}
