package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/mahmoud-eskandari/gojen/consts"
	"github.com/mahmoud-eskandari/gojen/internal/config"
	"github.com/mahmoud-eskandari/gojen/internal/version"
)

var Tables map[string]Table

func Generate() error {
	err := checkDir()
	if err != nil {
		return err
	}

	t1 := template.New("t1").Funcs(template.FuncMap{
		"replace":         strings.ReplaceAll,
		"camel":           strcase.ToCamel,
		"camelDefault":    camelDefault,
		"goType":          goType,
		"structTag":       structTag,
		"structTagExcept": structTagExcept,
		"imports":         SetImports,
		"join":            strings.Join,
		"isPrimary": func(table string, colName string) bool {
			for _, pk := range Tables[table].Primaries {
				if pk.Name == colName {
					return true
				}
			}
			return false
		},
		"joinCols": func(cols []Column, template, delimiter string) string {
			out := []string{}
			for _, col := range cols {
				if strings.Contains(template, "%") {
					out = append(out, fmt.Sprintf(template, col.Name))
				} else {
					out = append(out, template)
				}
			}
			return strings.Join(out, delimiter)
		},
		"joinCamelCols": func(cols []Column, template, delimiter string) string {
			out := []string{}
			for _, col := range cols {
				if strings.Contains(template, "%") {
					out = append(out, fmt.Sprintf(template, strcase.ToCamel(col.Name)))
				}
			}
			return strings.Join(out, delimiter)
		},
		"singular": func(name string) string {
			p := pluralize.NewClient()
			return p.Singular(name)
		},
		"version": func() string {
			return version.Ver
		},
		"datetime": func() string {
			return time.Now().Format(time.RFC1123)
		},
		"NonPrimaryCols": NonPrimaryCols,
		"pkg": func() string {
			return config.Conf.Pkg
		},
		"table": func(tableName string) Table {
			return Tables[tableName]
		},
	})

	t1, err = t1.Parse(consts.StructTemplate)
	if err != nil {
		return err
	}
	t1 = template.Must(t1.Parse(consts.StructTemplate))

	Tables = GetTables()
	for _, table := range Tables {
		filepath := path.Join(config.Conf.Dir, table.Name+".generated.go")
		if !config.Conf.Replace {
			_, err := os.Stat(filepath)
			if err == nil {
				log.Printf("[%s] regenerate ignored because `replace` is disabled in config", filepath)
				continue
			}
		}

		os.Remove(filepath)
		file, err := os.Create(filepath)
		if err != nil {
			panic(err)
		}

		var out []byte
		buf := bytes.NewBuffer(out)

		table.Prepare()
		err = t1.Execute(buf, table)
		if err != nil {
			log.Print(exec.ErrDot)
			file.Close()
			continue
		}

		out, _ = format.Source(out)
		file.Write(out)
		file.Close()
	}

	return nil
}

func checkDir() error {
	info, err := os.Stat(config.Conf.Dir)
	if err != nil {
		err = os.Mkdir(config.Conf.Dir, os.ModePerm)
		return err
	}

	if !info.IsDir() {
		return errors.New(config.Conf.Dir + " is not a dir.")
	}
	return nil
}
