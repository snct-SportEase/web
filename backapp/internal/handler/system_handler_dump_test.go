package handler

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSQLDump(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = ?
		ORDER BY CASE WHEN table_type = 'VIEW' THEN 1 ELSE 0 END, table_name
	`)).
		WithArgs("sportease").
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "table_type"}).
			AddRow("users", "BASE TABLE").
			AddRow("class_scores", "VIEW"))

	mock.ExpectQuery(regexp.QuoteMeta("SHOW CREATE TABLE `users`")).
		WillReturnRows(sqlmock.NewRows([]string{"Table", "Create Table"}).
			AddRow("users", "CREATE TABLE `users` (`id` int NOT NULL, `email` varchar(255) DEFAULT NULL)"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).
			AddRow(1, "s2301059@sendai-nct.jp").
			AddRow(2, "quote'test@example.com").
			AddRow(3, nil))

	mock.ExpectQuery(regexp.QuoteMeta("SHOW CREATE TABLE `class_scores`")).
		WillReturnRows(sqlmock.NewRows([]string{"View", "Create View", "character_set_client", "collation_connection"}).
			AddRow("class_scores", "CREATE VIEW `class_scores` AS SELECT `users`.`id` AS `id` FROM `users`", "utf8mb4", "utf8mb4_unicode_ci"))

	var dump bytes.Buffer
	err = writeSQLDump(context.Background(), db, "sportease", &dump)
	require.NoError(t, err)

	dumpText := dump.String()
	assert.Contains(t, dumpText, "SET FOREIGN_KEY_CHECKS=0;")
	assert.Contains(t, dumpText, "DROP VIEW IF EXISTS `class_scores`;")
	assert.Contains(t, dumpText, "DROP TABLE IF EXISTS `users`;")
	assert.Contains(t, dumpText, "CREATE TABLE `users`")
	assert.Contains(t, dumpText, "INSERT INTO `users` (`id`, `email`) VALUES ('1', 's2301059@sendai-nct.jp');")
	assert.Contains(t, dumpText, "INSERT INTO `users` (`id`, `email`) VALUES ('2', 'quote\\'test@example.com');")
	assert.Contains(t, dumpText, "INSERT INTO `users` (`id`, `email`) VALUES ('3', NULL);")
	assert.Contains(t, dumpText, "CREATE VIEW `class_scores`")
	assert.Contains(t, dumpText, "SET FOREIGN_KEY_CHECKS=1;")
	assert.NoError(t, mock.ExpectationsWereMet())
}
