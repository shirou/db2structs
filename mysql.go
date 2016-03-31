package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type MySQL struct{}

func (m MySQL) getSchema(config Configuration) []ColumnSchema {
	scheme := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema", config.DbUser, config.DbPassword, config.DbHost, config.DbPort)

	conn, err := sql.Open("mysql", scheme)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	q := `SELECT TABLE_NAME, COLUMN_NAME, IS_NULLABLE, DATA_TYPE, 
		CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, NUMERIC_SCALE, COLUMN_TYPE, 
		COLUMN_KEY FROM COLUMNS WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME, ORDINAL_POSITION`

	rows, err := conn.Query(q, config.DbName)
	if err != nil {
		log.Fatal(err)
	}
	columns := []ColumnSchema{}
	for rows.Next() {
		cs := ColumnSchema{}
		err := rows.Scan(&cs.TableName, &cs.ColumnName, &cs.IsNullable, &cs.DataType,
			&cs.CharacterMaximumLength, &cs.NumericPrecision, &cs.NumericScale,
			&cs.ColumnType, &cs.ColumnKey)
		if err != nil {
			log.Fatal(err)
		}
		columns = append(columns, cs)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return columns
}

func (m MySQL) goType(col *ColumnSchema) (string, string, error) {
	requiredImport := ""
	if col.IsNullable == "YES" {
		requiredImport = "database/sql"
	}
	var gt string = ""
	switch col.DataType {
	case "char", "varchar", "enum", "text", "longtext", "mediumtext", "tinytext":
		if col.IsNullable == "YES" {
			gt = "sql.NullString"
		} else {
			gt = "string"
		}
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		gt = "[]byte"
	case "date", "time", "datetime", "timestamp":
		gt, requiredImport = "time.Time", "time"
	case "tinyint", "smallint", "int", "mediumint", "bigint":
		if col.IsNullable == "YES" {
			gt = "sql.NullInt64"
		} else {
			gt = "int64"
		}
	case "float", "decimal", "double":
		if col.IsNullable == "YES" {
			gt = "sql.NullFloat64"
		} else {
			gt = "float64"
		}
	}
	if gt == "" {
		n := col.TableName + "." + col.ColumnName
		return "", "", errors.New("No compatible datatype (" + col.DataType + ") for " + n + " found")
	}
	return gt, requiredImport, nil
}

func (m MySQL) getEnvMap() map[string]string {
	return map[string]string{
		EnvHostKey:     "MYSQL_HOST",
		EnvPortKey:     "MYSQL_PORT",
		EnvDataBaseKey: "MYSQL_DATABASE",
		EnvUserKey:     "MYSQL_USER",
		EnvPasswordKey: "MYSQL_PASSWORD",
	}
}
