package generator

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/a-dminator/genna/database"
	"github.com/a-dminator/genna/model"

	"go.uber.org/zap"
)

func TestDo(t *testing.T) {
	user := model.Table{
		Schema: model.PublicSchema,
		Name:   "users",
		Columns: []model.Column{
			{
				Name:       "userId",
				Type:       model.TypeInt8,
				IsPK:       true,
				IsNullable: false,
			},
			{
				Name:       "email",
				Type:       model.TypeVarchar,
				IsNullable: true,
			},
			{
				Name:       "locationId",
				Type:       model.TypeInt8,
				IsNullable: false,
				IsFK:       true,
			},
			{
				Name: "companyId",
				Type: model.TypeInt8,
				IsFK: true,
			},
			{
				Name: "createdAt",
				Type: model.TypeTimestamp,
			},
		},
		Relations: []model.Relation{
			{
				Type:          model.HasOne,
				SourceSchema:  model.PublicSchema,
				SourceTable:   "users",
				SourceColumns: []string{"locationId"},
				TargetSchema:  "geo",
				TargetTable:   "locations",
				TargetColumns: []string{"locationId"},
			},
			{
				Type:          model.HasOne,
				SourceSchema:  model.PublicSchema,
				SourceTable:   "users",
				SourceColumns: []string{"companyId"},
				TargetSchema:  model.PublicSchema,
				TargetTable:   "companies",
				TargetColumns: []string{"companyId"},
			},
		},
	}

	company := model.Table{
		Schema: model.PublicSchema,
		Name:   "companies",
		Columns: []model.Column{
			{
				Name:       "companyId",
				Type:       model.TypeInt8,
				IsPK:       true,
				IsNullable: false,
			},
			{
				Name:       "title",
				Type:       model.TypeVarchar,
				IsNullable: true,
			},
		},
	}

	location := model.Table{
		Schema: "geo",
		Name:   "locations",
		Columns: []model.Column{
			{
				Name:       "locationId",
				Type:       model.TypeInt8,
				IsPK:       true,
				IsNullable: false,
			},
			{
				Name:       "title",
				Type:       model.TypeVarchar,
				IsNullable: true,
			},
		},
	}

	lang := model.Table{
		Schema: "geo",
		Name:   "languages",
		Columns: []model.Column{
			{
				Name:       "languageId",
				Type:       model.TypeInt8,
				IsPK:       true,
				IsNullable: false,
			},
			{
				Name:       "title",
				Type:       model.TypeVarchar,
				IsNullable: true,
			},
		},
	}

	unused := model.Table{
		Schema: model.PublicSchema,
		Name:   "unused",
		Columns: []model.Column{
			{
				Name:       "unusedId",
				Type:       model.TypeInt8,
				IsPK:       true,
				IsNullable: false,
			},
			{
				Name:       "title",
				Type:       model.TypeVarchar,
				IsNullable: true,
			},
		},
	}

	_, filename, _, _ := runtime.Caller(0)

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.Encoding = "console"
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	// just for test
	generator := NewGenerator(Options{
		Package:   "test", // model.DefaultPackage,
		Tables:    []string{"public.users", "geo.*"},
		FollowFKs: true,
		Output:    path.Dir(filename) + "/../test/model.go",
		KeepPK:    false, // try true
		NoDiscard: false, // try true
	}, logger)

	_, err = generator.Process([]model.Table{unused, user, company, location, lang})
	fmt.Print(err)
}

func TestLive(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	url := `postgres://genna:genna@localhost:5432/genna?sslmode=disable`
	options := Options{
		Package:      model.DefaultPackage,
		Tables:       []string{"public.*"},
		FollowFKs:    true,
		Output:       path.Dir(filename) + "/../test/model.go",
		KeepPK:       false, // try true
		NoDiscard:    false, // try true
		WithSearch:   true,
		StrictSearch: false,
	}

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.Encoding = "console"
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	db, err := database.NewDatabase(url, logger)
	if err != nil {
		panic(err)
	}

	store := database.NewStore(db)
	tables, err := store.Tables(model.Schemas(options.Tables))
	if err != nil {
		panic(err)
	}

	genna := NewGenerator(options, logger)

	if _, err := genna.Process(tables); err != nil {
		panic(err)
	}
}

func Test_addSuffix(t *testing.T) {
	type args struct {
		filename string
		suffix   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should add suffix to normal path",
			args: args{
				filename: "/some/dir/file.ext",
				suffix:   "_suf",
			},
			want: "/some/dir/file_suf.ext",
		},
		{
			name: "should add suffix to root file path",
			args: args{
				filename: "/file.ext",
				suffix:   "_suf",
			},
			want: "/file_suf.ext",
		},
		{
			name: "should add suffix to only file path",
			args: args{
				filename: "file.ext",
				suffix:   "_suf",
			},
			want: "file_suf.ext",
		},
		{
			name: "should add suffix to path without ext",
			args: args{
				filename: "/some/dir/file",
				suffix:   "_suf",
			},
			want: "/some/dir/file_suf",
		},
		{
			name: "should add suffix to file without ext",
			args: args{
				filename: "file",
				suffix:   "_suf",
			},
			want: "file_suf",
		},
		{
			name: "should add suffix to path with dots",
			args: args{
				filename: "/some.dir/fi.le.ext",
				suffix:   "_suf",
			},
			want: "/some.dir/fi.le_suf.ext",
		},
		{
			name: "should add suffix to path with dots and without ext",
			args: args{
				filename: "/some.dir/file",
				suffix:   "_suf",
			},
			want: "/some.dir/file_suf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addSuffix(tt.args.filename, tt.args.suffix); got != tt.want {
				t.Errorf("addSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}
