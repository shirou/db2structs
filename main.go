package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

var config Configuration

type dbType interface {
	getSchema(config Configuration) []ColumnSchema
	goType(col *ColumnSchema) (string, string, error)
}

type Configuration struct {
	DbType     string `json:"db_type"`
	DbUser     string `json:"db_user"`
	DbPassword string `json:"db_password"`
	DbName     string `json:"db_name"`
	DbHost     string `json:"db_host"`
	DbPort     int    `json:"db_port"`
	OutputFile string `json:"output_file"`
	// PkgName gives name of the package using the stucts
	PkgName string `json:"pkg_name"`
	// SQLTag produces tags commonly used to match database field names with Go struct members
	SQLTag string `json:"sql_tag"`
	// StructTag produces a tag to each struct
	StructTag string `json:"struct_tag"`
}

type ColumnSchema struct {
	TableName              string
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnKey              string
}

func getOutput(config Configuration, db dbType, schemas []ColumnSchema) ([]byte, error) {
	currentTable := ""
	var neededImports []string

	// First, get body text into var out
	out := ""
	for _, cs := range schemas {
		if cs.TableName != currentTable {
			if currentTable != "" {
				out += "}\n\n"
			}
			out += "// " + formatName(cs.TableName) + "\n"
			if config.StructTag != "" {
				out += "// " + config.StructTag + "\n"
			}
			out += "type " + formatName(cs.TableName) + " struct{\n"
		}

		goType, requiredImport, err := db.goType(&cs)
		if requiredImport != "" {
			neededImports = append(neededImports, requiredImport)
		}

		if err != nil {
			return []byte{}, err
		}
		out += "\t" + formatName(cs.ColumnName) + " " + goType
		tags := []string{"column:" + cs.ColumnName}
		switch cs.ColumnKey {
		case "PRI":
			tags = append(tags, "primary_key")
		case "UNI":
			tags = append(tags, "unique")
		}
		if cs.IsNullable != "YES" {
			tags = append(tags, "not null")
		}

		if len(config.SQLTag) > 0 {
			out += "\t`" + fmt.Sprintf(`%s:"%s"`, config.SQLTag, strings.Join(tags, ";")) + "`"
		}

		out += "\n"
		currentTable = cs.TableName

	}
	out = out + "}"

	// Build the header section
	headerTmpl := "package %s \n\n %s"
	imports := ""

	if len(neededImports) > 0 {
		imports = "import (\n"
		for _, imp := range neededImports {
			imports += "\t\"" + imp + "\"\n"
		}
		imports += ")\n\n"
	}
	header := fmt.Sprintf(headerTmpl, config.PkgName, imports)

	return format.Source([]byte(header + out))
}

func writeStructs(config Configuration, output []byte) error {
	if config.OutputFile == "" { // Output stdout if not specified
		fmt.Println(string(output))
		return nil
	}

	file, err := os.Create(config.OutputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = fmt.Fprint(file, string(output))
	return err
}

func formatName(name string) string {
	parts := strings.Split(name, "_")
	newName := ""
	for _, p := range parts {
		if len(p) < 1 {
			continue
		}
		newName = newName + strings.Replace(p, string(p[0]), strings.ToUpper(string(p[0])), 1)
	}

	newName = strings.Replace(newName, "Id", "ID", -1) // for golint
	// If a first charactor of the table is number, add "A" to the top
	if unicode.IsNumber(rune(newName[0])) {
		newName = "A" + newName
	}

	return newName
}

var configFile = flag.String("json", "", "Config file")

func usage() {
	fmt.Printf("Usage of %s:\n  -json <JSON file>\n", os.Args[0])
	fmt.Println(`  or use these environmental variables.
MYSQL_HOST
MYSQL_PORT
MYSQL_DATABASE
MYSQL_USER
MYSQL_PASSWORD
`)
}

func overrideByEnv() error {
	v, ok := os.LookupEnv("MYSQL_HOST")
	if ok {
		config.DbHost = v
	}
	v, ok = os.LookupEnv("MYSQL_PORT")
	if ok {
		p, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("parse error MYSQL_PORT, %s", err)
		}
		config.DbPort = p
	}
	v, ok = os.LookupEnv("MYSQL_DATABASE")
	if ok {
		config.DbName = v
	}
	v, ok = os.LookupEnv("MYSQL_USER")
	if ok {
		config.DbUser = v
	}
	v, ok = os.LookupEnv("MYSQL_PASSWORD")
	if ok {
		config.DbPassword = v
	}
	return nil
}

// NewDB returns DBtype
func NewDB(dbType string) (dbType, error) {
	return MySQL{}, nil
}

func main() {
	flag.Parse()
	if *configFile != "" {
		f, err := os.Open(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = json.NewDecoder(f).Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	}
	err := overrideByEnv()
	if err != nil {
		log.Fatal(err)
	}
	empty := Configuration{}
	if config == empty {
		usage()
		os.Exit(0)
	}
	if config.DbType == "" {
		config.DbType = "mysql"
	}

	db, err := NewDB(config.DbType)
	if err != nil {
		log.Fatal(err)
	}

	columns := db.getSchema(config)
	output, err := getOutput(config, db, columns)
	if err != nil {
		log.Fatal(err)
	}

	err = writeStructs(config, output)
	if err != nil {
		log.Fatal(err)
	}
}