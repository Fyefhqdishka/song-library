package repository_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Fyefhqdishka/song-library/repositories"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"regexp"
	"testing"
)

func TestUpdateSongText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок базы данных: %s", err)
	}
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := repositories.NewSongRepository(db, logger)

	expectedQuery := regexp.QuoteMeta(`UPDATE songs SET text = $1 WHERE song_title ILIKE $2`)

	mock.ExpectExec(expectedQuery).
		WithArgs("новый текст", "Тестовая песня").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateSongText("Тестовая песня", "новый текст")
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %s", err)
	}
}
