package generator

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/mahmoud-eskandari/gojen/consts"
	"github.com/mahmoud-eskandari/gojen/internal/config"
)

func structTag(Column string) string {
	out := ""
	for i := range config.Conf.Tags {
		if i > 0 {
			out += " "
		}
		out += config.Conf.Tags[i] + ":\"" + Column + "\""
	}

	if out != "" {
		out = "`" + out + "`"
	}
	return out
}

func structTagExcept(Column string, Except string) string {
	out := ""
	for i := range config.Conf.Tags {
		if out != "" {
			out += " "
		}
		col := Column
		if config.Conf.Tags[i] == Except {
			col = "-"
		}
		out += config.Conf.Tags[i] + ":\"" + col + "\""
	}

	if out != "" {
		out = "`" + out + "`"
	}
	return out
}

func goType(ColumnType string) string {
	if t, ok := consts.MysqlDicMp[ColumnType]; ok {
		return t
	}
	return "interface{}"
}

func camelDefault(s string, def interface{}) string {
	camel := strcase.ToCamel(s)
	if camel == "" {
		return fmt.Sprintf("%v", def)
	}
	return camel
}

func SetImports(imports []string) string {
	if len(imports) == 0 {
		return ""
	}
	if len(imports) == 1 {
		return `"` + imports[0] + `"`
	}

	out := ""
	visited := make(map[string]bool)
	for _, imp := range imports {
		if ok := visited[imp]; ok {
			continue
		}
		visited[imp] = true
		out += "	\"" + imp + "\"\n"
	}
	return out
}

func NonPrimaryCols(t Table) []Column {
	out := []Column{}
	for _, col := range t.Columns {
		if COlContains(t.Primaries, col) < 0 {
			out = append(out, col)
		}
	}
	return out
}

func (t *Table) Prepare() {
	if t.PrimaryColType == "string" {
		t.Imports = append(t.Imports, "strconv")
	}
}

// COlContains check heyStack contains needle and return needle`s position
func COlContains(heyStack []Column, needle Column) int {
	for i, item := range heyStack {
		if item.Name == needle.Name {
			return i
		}
	}
	return -1
}

// StringContains check heyStack contains needle and return needle`s position
func StringContains(heyStack []string, needle string) int {
	for i, item := range heyStack {
		if item == needle {
			return i
		}
	}
	return -1
}

func RemoveItem(s []Relation, i int) []Relation {
	if i < 0 || i > len(s)-1 {
		return s
	}

	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
