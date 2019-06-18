package generator

import (
	"fmt"
	"html/template"

	"github.com/a-dminator/genna/model"
)

// Stores package info
type templatePackage struct {
	Package string

	HasImports bool
	Imports    []string

	Models []templateTable

	HasSearchImports bool
	SearchImports    []string
}

// newMultiPackage creates a package with multiple models
func newTemplatePackage(tables []model.Table, options Options) templatePackage {
	mImports := make([]string, 0)
	sImports := make([]string, 0)

	models := make([]templateTable, len(tables))
	for i, table := range tables {
		mImports = append(mImports, table.Imports()...)
		sImports = append(sImports, table.SearchImports()...)

		models[i] = newTemplateTable(table, options)
		models[i].uniqualizeFields()
	}

	mImports = model.UniqStrings(mImports)
	sImports = model.UniqStrings(sImports)

	return templatePackage{
		Package:    options.Package,
		HasImports: len(mImports) > 0,
		Imports:    mImports,

		Models: models,

		HasSearchImports: options.StrictSearch && len(sImports) > 0,
		SearchImports:    sImports,
	}
}

// stores struct info
type templateTable struct {
	StructName string
	StructTag  template.HTML

	TableName string

	WithAlias  bool
	TableAlias string
	JoinAlias  string

	Columns []templateColumn

	HasRelations bool
	Relations    []templateRelation

	SearchStructName string
}

func newTemplateTable(table model.Table, options Options) templateTable {
	if table.HasMultiplePKs() {
		options.KeepPK = true
	}

	columns := make([]templateColumn, len(table.Columns))
	for i, column := range table.Columns {
		columns[i] = newTemplateColumn(column, options)
	}

	relations := make([]templateRelation, len(table.Relations))
	for i, relation := range table.Relations {
		relations[i] = newTemplateRelation(relation)
	}

	return templateTable{
		StructName: table.ModelName(),
		StructTag:  template.HTML(fmt.Sprintf("`%s`", table.TableNameTag(options.View, options.NoAlias, options.NoDiscard))),

		TableName: table.TableName(false),

		WithAlias:  !options.NoAlias,
		TableAlias: table.Alias(),
		JoinAlias:  table.JoinAlias(),

		Columns: columns,

		HasRelations: len(relations) > 0,
		Relations:    relations,

		SearchStructName: table.SearchModelName(),
	}
}

// stores column info
type templateColumn struct {
	FieldName    string
	FieldDBName  string
	FieldType    string
	FieldTag     template.HTML
	FieldComment template.HTML

	IsSearchable    bool
	SearchFieldType string
}

func newTemplateColumn(column model.Column, options Options) templateColumn {
	return templateColumn{
		FieldName:    column.StructFieldName(options.KeepPK),
		FieldDBName:  column.Name,
		FieldType:    column.StructFieldType(),
		FieldTag:     template.HTML(fmt.Sprintf("`%s`", column.StructFieldTag(options.SoftDelete))),
		FieldComment: template.HTML(column.Comment()),

		IsSearchable:    column.IsSearchable(),
		SearchFieldType: column.SearchFieldType(options.StrictSearch),
	}
}

// stores relation info
type templateRelation struct {
	FieldName    string
	FieldType    string
	FieldTag     template.HTML
	FieldComment template.HTML
}

func newTemplateRelation(relation model.Relation) templateRelation {
	return templateRelation{
		FieldName:    relation.StructFieldName(),
		FieldType:    relation.StructFieldType(),
		FieldTag:     template.HTML(fmt.Sprintf("`%s`", relation.StructFieldTag())),
		FieldComment: template.HTML(relation.Comment()),
	}
}

func (t templateTable) uniqualizeFields() {
	index := map[string]bool{}

	for i, column := range t.Columns {
		fieldName := column.FieldName

		if _, ok := index[fieldName]; !ok {
			index[fieldName] = true
			continue
		}

		suffix := 1
	couter:
		for {
			fieldName = fmt.Sprintf("%s%d", column.FieldName, suffix)

			for _, col := range t.Columns {
				if col.FieldName == fieldName {
					suffix++
					continue couter
				}
			}
			t.Columns[i].FieldName = fieldName
			break
		}
	}

	for i, relation := range t.Relations {
		fieldName := relation.FieldName

		if _, ok := index[fieldName]; !ok {
			index[fieldName] = true
			continue
		}

		suffix := 0

	router:
		for {
			if suffix == 0 {
				fieldName = fmt.Sprintf("%s%s", relation.FieldName, model.NonUniqSuffix)
			} else {
				fieldName = fmt.Sprintf("%s%s%d", relation.FieldName, model.NonUniqSuffix, suffix)
			}

			for _, col := range t.Columns {
				if col.FieldName == fieldName {
					suffix++
					continue router
				}
			}

			for _, rel := range t.Relations {
				if rel.FieldName == fieldName {
					suffix++
					continue router
				}
			}

			t.Relations[i].FieldName = fieldName
			break
		}
	}
}
