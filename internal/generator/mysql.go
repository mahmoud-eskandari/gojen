package generator

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/mahmoud-eskandari/gojen/internal/config"
	"github.com/mahmoud-eskandari/gojen/internal/repo"
)

func GetTables() map[string]Table {

	rows, err := repo.DB.Query(`SELECT 
	TABLES.TABLE_NAME,
	JSON_ARRAYAGG(JSON_OBJECT('name',COLUMNS.COLUMN_NAME,'data_type',COLUMNS.DATA_TYPE,'column_type',COLUMNS.COLUMN_TYPE,'nullable',IF(COLUMNS.IS_NULLABLE = 'NO',0,1),'length',IFNULL(COLUMNS.CHARACTER_MAXIMUM_LENGTH,-1),'default',COLUMNS.COLUMN_DEFAULT,'extra',COLUMNS.EXTRA)) AS cols
	FROM
	TABLES
	INNER JOIN COLUMNS ON COLUMNS.TABLE_SCHEMA = TABLES.TABLE_SCHEMA AND COLUMNS.TABLE_NAME = TABLES.TABLE_NAME
	WHERE TABLES.TABLE_SCHEMA=?
	GROUP BY TABLES.TABLE_SCHEMA, TABLES.TABLE_NAME`, config.Conf.DB.Schema)
	if err != nil {
		panic(err)
	}

	var out = make(map[string]Table)
	p := pluralize.NewClient()

	for rows.Next() {
		var ct Table
		var cols string
		rows.Scan(&ct.Name, &cols)
		ct.SingularName = ct.Name

		ct.SingularName = strcase.ToCamel(p.Singular(ct.Name))

		err = json.Unmarshal([]byte(cols), &ct.Columns)
		if err != nil {
			panic(err)
		}

		//Parse ENUMs
		for i := range ct.Columns {
			if ct.Columns[i].DataType == "enum" {
				str := strings.ReplaceAll(ct.Columns[i].ColumnType, "enum(", "")
				str = strings.ReplaceAll(str, ")", "")
				str = strings.ReplaceAll(str, "'", "")
				ct.Columns[i].Enum = strings.Split(str, ",")
			}
			ct.SetGoDataType(i)
		}
		//get relations
		ct.Getchilds()
		//get parent relations
		ct.GetParents()
		//get indexes
		ct.GetIndexes()

		out[ct.Name] = ct
	}

	//put relation table addresses
	for tableName, table := range out {
		visited := []string{}
		for relId := 0; relId < len(table.LRelations); relId++ {
			rel := table.LRelations[relId]
			key := rel.FromCol + rel.ToCol + rel.RefTable
			if StringContains(visited, key) > -1 {
				table.LRelations = RemoveItem(table.LRelations, relId)
				out[tableName] = table
				relId--
				continue
			}
			visited = append(visited, key)
			RelTable, ok := out[rel.RefTable]
			if !ok {
				panic(rel.RefTable + " dosen't exists in table list that is needed for relation of " + tableName)
			}
			out[tableName].LRelations[relId].Table = &RelTable
		}
	}
	return out
}

func (t *Table) SetGoDataType(i int) {
	dt := goType(t.Columns[i].DataType)

	if t.Columns[i].Nullable == 1 && dt != "[]byte" {
		t.Imports = append(t.Imports, "database/sql")
		dt = makeNullable(dt)
	}
	if strings.Contains(dt, "time.") {
		t.Imports = append(t.Imports, "time")
	}

	t.Columns[i].GoDataType = dt
}

func (t *Table) GetIndexes() {
	rows, err := repo.DB.Query(`SELECT DISTINCT INDEX_NAME,COLUMN_NAME FROM INFORMATION_SCHEMA.STATISTICS WHERE  NON_UNIQUE=0 AND TABLE_SCHEMA = ? AND TABLE_NAME=?`,
		config.Conf.DB.Schema,
		t.Name)
	if err != nil {
		log.Printf("error on read indexes of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var indexName, columnName string
		rows.Scan(&indexName, &columnName)
		if indexName == "PRIMARY" {
			for _, col := range t.Columns {
				if col.Name == columnName {
					t.Primaries = append(t.Primaries, col)
					break
				}
			}
		} else {
			if t.UniqIndexes == nil {
				t.UniqIndexes = make(map[string][]Column)
			}
			for _, col := range t.Columns {
				if col.Name == columnName {
					if _, ok := t.UniqIndexes[indexName]; ok {
						t.UniqIndexes[indexName] = append(t.UniqIndexes[indexName], col)
					} else {
						t.UniqIndexes[indexName] = []Column{}
						t.UniqIndexes[indexName] = append(t.UniqIndexes[indexName], col)
					}

				}
			}
		}
	}

	if len(t.Primaries) == 1 {
		t.Primary = t.Primaries[0].Name
		t.PrimaryColType = t.Primaries[0].GoDataType
	}
}

func (t *Table) Getchilds() {
	rows, err := repo.DB.Query(`SELECT DISTINCT
	column_name AS column_name,
	referenced_table_name AS referenced_table_name,
	referenced_column_name AS referenced_column_name
	from INFORMATION_SCHEMA.KEY_COLUMN_USAGE where table_schema = ? AND REFERENCED_TABLE_NAME IS NOT NULL AND TABLE_NAME = ?`,
		config.Conf.DB.Schema,
		t.Name)
	if err != nil {
		log.Printf("error on read childs of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var rel Relation
		rows.Scan(&rel.FromCol, &rel.RefTable, &rel.ToCol)
		t.LRelations = append(t.LRelations, rel)
	}
}

func (t *Table) GetParents() {
	rows, err := repo.DB.Query(`SELECT DISTINCT
	column_name AS column_name,
	TABLE_NAME,
	referenced_column_name AS referenced_column_name
	from INFORMATION_SCHEMA.KEY_COLUMN_USAGE where REFERENCED_TABLE_SCHEMA = ? AND referenced_table_name = ? AND TABLE_NAME != ?`,
		config.Conf.DB.Schema,
		t.Name,
		t.Name)
	if err != nil {
		log.Printf("error on read parents of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var rel Relation
		rows.Scan(&rel.FromCol, &rel.RefTable, &rel.ToCol)
		t.RRelations = append(t.RRelations, rel)
	}
}

func makeNullable(ctype string) string {
	switch ctype {
	case "string":
		return "sql.NullString"
	case "[]byte":
		return "[]byte"
	case "time.Time":
		return "sql.NullTime"
	case "uint", "uint32", "uint8", "uint16":
		return "sql.NullInt32"
	case "int", "int32":
		return "sql.NullInt32"
	case "int8", "int16":
		return "sql.NullInt16"
	case "int64", "uint64":
		return "sql.NullInt64"
	case "float32", "float64":
		return "sql.NullFloat64"
	case "bool":
		return "sql.NullBool"
	}
	return "sql.NullString"
}
