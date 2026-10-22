package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestEncryptionDeletionLocksAppAndPreservesSelectedPolicy(t *testing.T) {
	for _, selected := range []int64{7, 0} {
		raw, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		db := sqlx.NewDb(raw, "sqlmock")
		InitTablePrefix("a_")
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT encryption FROM a_apps WHERE id=\? FOR UPDATE`).WithArgs(int64(1000)).WillReturnRows(sqlmock.NewRows([]string{"encryption"}).AddRow(selected))
		if selected == 7 {
			mock.ExpectRollback()
		} else {
			mock.ExpectExec(`DELETE FROM a_encrypts WHERE appid = \? AND id = \?`).WithArgs(int64(1000), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
		}
		err = NewEncryptRepo(db).Delete(context.Background(), 1000, 7)
		if selected == 7 && !errors.Is(err, ErrEncryptionInUse) || selected == 0 && err != nil {
			t.Fatal(err)
		}
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func TestAppUpdateRechecksPolicyUnderLock(t *testing.T) {
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	InitTablePrefix("a_")
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM a_apps WHERE id=\? FOR UPDATE`).WithArgs(int64(1000)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1000))
	mock.ExpectQuery(`SELECT id FROM a_encrypts WHERE appid=\? AND id=\? FOR UPDATE`).WithArgs(int64(1000), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	err = NewAppRepo(sqlx.NewDb(raw, "sqlmock")).Update(context.Background(), &App{ID: 1000, Encryption: 7})
	if !errors.Is(err, ErrEncryptionUnavailable) {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
