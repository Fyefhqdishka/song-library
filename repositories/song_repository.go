package repositories

import (
	"database/sql"
	"encoding/json"
	"github.com/Fyefhqdishka/song-library/models"
	"log/slog"
)

type SongMethods interface {
	AddSong(song *models.Song) error
	DeleteSong(groupName, songTitle string) error
	GetSong(song *models.GetSong, limit int, offset int) ([]*models.Song, error)
	GetSongVerses(songTitle string, page int, limit int) ([]string, error)
	UpdateSongText(songTitle, newText string) error
}

type SongRepository struct {
	DB     *sql.DB
	Logger *slog.Logger
}

func NewSongRepository(db *sql.DB, logger *slog.Logger) *SongRepository {
	return &SongRepository{db, logger}
}

// AddSong godoc
// @Summary 	Add a new song
// @Description Add a new song to the library
// @Tags 		songs
// @Accept 		json
// @Produce 	json
// @Param 		song body models.AddSongRequest true "Song data"
// @Success 	200 {object} models.Song
// @Failure 	400 {string} string "Bad Request"
// @Failure 	500 {string} string "Internal Server Error"
// @Router 		/songs [post]
func (m *SongRepository) AddSong(song *models.Song) error {
	m.Logger.Debug("Добавление новой песни")

	versesJSON, err := json.Marshal(song.Verses)
	if err != nil {
		m.Logger.Error("Ошибка при сериализации куплетов в JSON")
		return err
	}

	stmt := `INSERT INTO songs (group_name, song_title, release_date, text, link, verses) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = m.DB.Exec(stmt, song.GroupName, song.SongTitle, song.ReleaseDate, song.Text, song.Link, versesJSON)
	if err != nil {
		m.Logger.Error("Ошибка при добавлении новой песни")
		return err
	}
	m.Logger.Info("verses:", song.Verses)

	m.Logger.Debug("Песня успешно добавлена", "группа", song.GroupName, "название", song.SongTitle)
	return nil
}

// DeleteSong godoc
// @Summary 	Delete Song
// @Description Delete a song from the library
// @Tags 		songs
// @Accept 		json
// @Produce 	json
// @Param 		song body models.Song true "Song data"
// @Success 	200 {object} models.Song
// @Failure 	400 {string} string "Bad request"
// @Failure 	500 {string} string "Internal server error"
// @Router 		/songs [delete]
func (m *SongRepository) DeleteSong(groupName, songTitle string) error {
	m.Logger.Debug("Удаление песни", "группа", groupName, "название", songTitle)

	stmt := `DELETE FROM songs WHERE group_name=$1 AND song_title=$2`
	_, err := m.DB.Exec(stmt, groupName, songTitle)
	if err != nil {
		m.Logger.Error("Ошибка при удалении песни", "группа", groupName, "название", songTitle, "ошибка", err)
		return err
	}
	m.Logger.Debug("Песня успешно удалена", "группа", groupName, "название", songTitle)
	return nil
}

// GetSong godoc
// @Summary 	Get Songs
// @Description Get a list of songs from the library based on filters
// @Tags 		songs
// @Accept 		json
// @Produce 	json
// @Param 		groupName query string false "Filter by group name"
// @Param 		songTitle query string false "Filter by song title"
// @Param 		releaseDate query string false "Filter by release date (format: YYYY-MM-DD)"
// @Param 		text query string false "Filter by song text"
// @Param 		link query string false "Filter by song link"
// @Param 		limit query int false "Limit the number of results"
// @Param 		offset query int false "Offset for pagination"
// @Success 	200 {array} models.Song
// @Failure 	400 {string} string "Bad query"
// @Failure 	500 {string} string "Internal server error"
// @Router 		/songs [get]
func (m *SongRepository) GetSong(song *models.GetSong, limit, offset int) ([]*models.Song, error) {
	m.Logger.Debug("Вход в GetSong", "группа", song.GroupName, "название", song.SongTitle, "limit", limit, "offset", offset)

	stmt := `
	SELECT * 
		FROM songs
			WHERE (COALESCE($1, '') = '' OR group_name ILIKE '%' || $1 || '%')
  			AND (COALESCE($2, '') = '' OR song_title ILIKE '%' || $2 || '%')
  			AND (release_date = $3)
  			AND (COALESCE($4, '') = '' OR text ILIKE '%' || $4 || '%')
  			AND (COALESCE($5, '') = '' OR link ILIKE '%' || $5 || '%')
		ORDER BY release_date DESC
		LIMIT $6 OFFSET $7;`

	rows, err := m.DB.Query(stmt, song.GroupName, song.SongTitle, song.ReleaseDate, song.Text, song.Link, limit, offset)
	if err != nil {
		m.Logger.Error("Ошибка выполнения запроса на получение песен", "error", err)
		return nil, err
	}
	defer rows.Close()

	var songs []*models.Song
	for rows.Next() {
		var s models.Song
		if err := rows.Scan(&s.ID, &s.GroupName, &s.SongTitle, &s.ReleaseDate, &s.Text, &s.Link, &s.Verses); err != nil {
			m.Logger.Error("Ошибка при сканировании строки", "error", err)
			return nil, err
		}
		songs = append(songs, &s)
	}

	if err := rows.Err(); err != nil {
		m.Logger.Error("Ошибка после сканирования строк", "error", err)
		return nil, err
	}

	m.Logger.Debug("Выход из GetSong", "количество песен", len(songs))
	return songs, err
}

// GetSongVerses godoc
// @Summary 	Get song verses
// @Description Retrieves verses of a song by its title with pagination
// @Tags 		songs
// @Accept 		json
// @Produce 	json
// @Param 		songTitle query string true "Song title"
// @Param 		limit query int true "Limit of verses per page"
// @Param 		offset query int true "Offset number"
// @Success 	200 {array} string "Verses list"
// @Failure 	400 {string} string "Invalid page or limit number"
// @Failure 	500 {string} string "Error retrieving verses"
// @Router 		/songs/verses [get]
func (m *SongRepository) GetSongVerses(songTitle string, offset int, limit int) ([]string, error) {
	var verses []string
	stmt := `SELECT verses FROM songs WHERE song_title ILIKE $1 LIMIT $2 OFFSET $3;`

	m.Logger.Debug("Запрос к БД:", "SQL", stmt, "songTitle", songTitle, "limit", limit, "offset", offset)

	rows, err := m.DB.Query(stmt, songTitle, limit, offset)
	if err != nil {
		m.Logger.Error("Ошибка выполнения запроса на получение текста песни по куплетам", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var versesJson sql.NullString
		if err := rows.Scan(&versesJson); err != nil {
			m.Logger.Error("Ошибка при сканировании строки", "error", err)
			return nil, err
		}
		if versesJson.Valid {
			var tempVerses []string
			err = json.Unmarshal([]byte(versesJson.String), &tempVerses)
			if err != nil {
				return nil, err
			}
			verses = append(verses, tempVerses...)
		}
	}
	if err := rows.Err(); err != nil {
		m.Logger.Error("Ошибка после сканирования строк", "error", err)
		return nil, err
	}

	m.Logger.Debug("Выход из GetSongVerses", "Получена песня", "song_title:", songTitle)
	return verses, nil
}

// UpdateSongText godoc
// @Summary      Update Song Text
// @Description  Обновляет текст песни по song_title
// @Tags         songs
// @Accept       json
// @Produce      json
// @Param        song_title  query  string  true  "Название песни"
// @Param        text        body   string  true  "Новый текст песни"
// @Success      200  {string}  string  "Текст песни обновлен"
// @Failure      400  {string}  string  "Некорректный запрос"
// @Failure      500  {string}  string  "Ошибка на стороне сервера"
// @Router       /update-text [put]
func (m *SongRepository) UpdateSongText(songTitle, newText string) error {
	m.Logger.Debug("Начало обновлении текста песни:")
	stmt := `UPDATE songs SET text = $1 WHERE song_title ILIKE $2`

	_, err := m.DB.Exec(stmt, newText, songTitle)
	if err != nil {
		m.Logger.Error("Ошибка выполнения запроса на обновление текста песни", "error", err)
		return err
	}

	m.Logger.Debug("Поле text успешно обновлено для песни", "song_title", songTitle)
	return nil
}
