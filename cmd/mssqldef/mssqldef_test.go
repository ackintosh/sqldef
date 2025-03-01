// Integration test of mssqldef command.
//
// Test requirement:
//   - go command
//   - `sqlcmd -Usa -PPassw0rd` must succeed
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

const (
	applyPrefix     = "-- Apply --\n"
	nothingModified = "-- Nothing is modified --\n"
)

func TestMssqldefColumnLiteral(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE v (
		  v_integer integer NOT NULL,
		  v_text text,
		  v_smallmoney smallmoney,
		  v_money money,
		  v_datetimeoffset datetimeoffset(1),
		  v_datetime2 datetime2,
		  v_smalldatetime smalldatetime,
		  v_nchar nchar(30),
		  v_nvarchar nvarchar(30),
		  v_ntext ntext
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableQuotes(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE test_table (
		  id integer
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(
		"CREATE TABLE test_table (\n" +
			"  id integer\n" +
			");\n",
	)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTable(t *testing.T) {
	resetTestDatabase()

	createTable1 := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name text,
		  age integer
		);
		`,
	)
	createTable2 := stripHeredoc(`
		CREATE TABLE bigdata (
		  data bigint
		);
		`,
	)

	assertApplyOutput(t, createTable1+createTable2, applyPrefix+createTable1+createTable2)
	assertApplyOutput(t, createTable1+createTable2, nothingModified)

	assertApplyOutput(t, createTable1, applyPrefix+"DROP TABLE [dbo].[bigdata];\n")
	assertApplyOutput(t, createTable1, nothingModified)
}

func TestMssqldefCreateTableWithDefault(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  profile varchar(50) NOT NULL DEFAULT '',
		  default_int int default 20,
		  default_bool bit default 1,
		  default_numeric numeric(5) default 42.195,
		  default_fixed_char varchar(3) default 'JPN',
		  default_text text default ''
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableWithIDENTITY(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id integer PRIMARY KEY IDENTITY(1,1),
		  name text,
		  age integer
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableWithCLUSTERED(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id integer,
		  name text,
		  age integer,
		  CONSTRAINT PK_users PRIMARY KEY CLUSTERED (id)
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateView(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE [dbo].[users] (
		  id integer NOT NULL,
		  name text,
		  age integer
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createView := stripHeredoc(`
		CREATE VIEW [dbo].[view_users] AS select id from dbo.users where age = 1;
		`,
	)
	assertApplyOutput(t, createTable+createView, applyPrefix+createView)
	assertApplyOutput(t, createTable+createView, nothingModified)

	createView = stripHeredoc(`
		CREATE VIEW [dbo].[view_users] AS select id from dbo.users where age = 2;
		`,
	)
	dropView := stripHeredoc(`
		DROP VIEW [dbo].[view_users];
		`,
	)
	assertApplyOutput(t, createTable+createView, applyPrefix+dropView+createView)
	assertApplyOutput(t, createTable+createView, nothingModified)

	assertApplyOutput(t, "", applyPrefix+"DROP TABLE [dbo].[users];\n"+dropView)
}

func TestMssqldefAddColumn(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id BIGINT NOT NULL PRIMARY KEY
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id BIGINT NOT NULL PRIMARY KEY,
		  name varchar(40)
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+"ALTER TABLE [dbo].[users] ADD [name] varchar(40);\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefAddColumnWithIDENTITY(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id BIGINT NOT NULL PRIMARY KEY
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id BIGINT NOT NULL PRIMARY KEY,
		  membership_id int IDENTITY(1,1)
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+"ALTER TABLE [dbo].[users] ADD [membership_id] int IDENTITY(1,1);\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropColumn(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY,
		  name varchar(20)
		);`,
	)
	assertApply(t, createTable)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+"ALTER TABLE [dbo].[users] DROP COLUMN [name];\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropColumnWithDefaultConstraint(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY,
		  name varchar(20) CONSTRAINT df_name DEFAULT NULL
		);`,
	)
	assertApply(t, createTable)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+"ALTER TABLE [dbo].[users] DROP CONSTRAINT [df_name];\n"+"ALTER TABLE [dbo].[users] DROP COLUMN [name];\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropColumnWithDefault(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY,
		  name varchar(20) DEFAULT NULL
		);`,
	)
	assertApply(t, createTable)

	// extract name of default constraint from sql server
	out, err := execute("sqlcmd", "-Usa", "-PPassw0rd", "-dmssqldef_test", "-h", "-1", "-Q", stripHeredoc(`
		SELECT OBJECT_NAME(c.default_object_id) FROM sys.columns c WHERE c.object_id = OBJECT_ID('dbo.users', 'U') AND c.default_object_id != 0;
		`,
	))
	if err != nil {
		t.Error("failed to extract default object id")
	}
	dfConstraintName := strings.Replace((strings.Split(out, "\n")[0]), " ", "", -1)
	dropConstraint := fmt.Sprintf("ALTER TABLE [dbo].[users] DROP CONSTRAINT [%s];\n", dfConstraintName)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+dropConstraint+"ALTER TABLE [dbo].[users] DROP COLUMN [name];\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropColumnWithPK(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20) DEFAULT NULL,
			CONSTRAINT pk_id PRIMARY KEY (id)
		);`,
	)
	assertApply(t, createTable)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  name varchar(20) DEFAULT NULL
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+"ALTER TABLE [dbo].[users] DROP CONSTRAINT [pk_id];\n"+"ALTER TABLE [dbo].[users] DROP COLUMN [id];\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableAddPrimaryKey(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20)
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY,
		  name varchar(20)
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+
		"ALTER TABLE [dbo].[users] ADD primary key CLUSTERED ([id]);\n",
	)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableAddPrimaryKeyConstraint(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20)
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		  CONSTRAINT [pk_users] PRIMARY KEY CLUSTERED ([id])
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+
		"ALTER TABLE [dbo].[users] ADD CONSTRAINT [pk_users] primary key CLUSTERED ([id]);\n",
	)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropPrimaryKey(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL PRIMARY KEY,
		  name varchar(20)
		);`,
	)
	assertApply(t, createTable)

	// extract name of primary key constraint from sql server
	out, err := execute("sqlcmd", "-Usa", "-PPassw0rd", "-dmssqldef_test", "-h", "-1", "-Q", stripHeredoc(`
		SELECT kc.name FROM sys.key_constraints kc WHERE kc.parent_object_id=OBJECT_ID('users', 'U') AND kc.[type]='PK';
		`,
	))
	if err != nil {
		t.Error("failed to extract primary key id")
	}
	pkConstraintName := strings.Replace((strings.Split(out, "\n")[0]), " ", "", -1)
	dropConstraint := fmt.Sprintf("ALTER TABLE [dbo].[users] DROP CONSTRAINT [%s];\n", pkConstraintName)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20)
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+dropConstraint)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropPrimaryKeyConstraint(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
			CONSTRAINT [pk_users] PRIMARY KEY ([id])
		);`,
	)
	assertApply(t, createTable)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20)
		);`,
	)
	assertApplyOutput(t, createTable, applyPrefix+"ALTER TABLE [dbo].[users] DROP CONSTRAINT [pk_users];\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableWithIndexOption(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		    INDEX [ix_users_id] UNIQUE CLUSTERED ([id]) WITH (
		      PAD_INDEX = ON,
		      FILLFACTOR = 10,
		      IGNORE_DUP_KEY = ON,
		      STATISTICS_NORECOMPUTE = ON,
		      STATISTICS_INCREMENTAL = OFF,
		      ALLOW_ROW_LOCKS = ON,
		      ALLOW_PAGE_LOCKS = ON
		  )
		);
		`)

	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTablePrimaryKeyWithIndexOption(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		  CONSTRAINT [pk_users] PRIMARY KEY CLUSTERED ([id]) WITH (
		    PAD_INDEX = OFF,
		    STATISTICS_NORECOMPUTE = OFF,
		    IGNORE_DUP_KEY = OFF,
		    ALLOW_ROW_LOCKS = ON,
		    ALLOW_PAGE_LOCKS = ON
		  )
		);
		`)

	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableAddIndex(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20)
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		  INDEX [ix_users_id] UNIQUE CLUSTERED ([id]) WITH (
		    PAD_INDEX = ON,
		    FILLFACTOR = 10,
		    STATISTICS_NORECOMPUTE = ON
		  )
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+
		"CREATE UNIQUE CLUSTERED INDEX [ix_users_id] ON [dbo].[users] ([id]) WITH (pad_index = ON, fillfactor = 10, statistics_norecompute = ON);\n",
	)
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableDropIndex(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		  INDEX [ix_users_id] UNIQUE CLUSTERED ([id]) WITH (
		    PAD_INDEX = ON,
		    FILLFACTOR = 10,
		    STATISTICS_NORECOMPUTE = ON
		  )
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20)
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+"DROP INDEX [ix_users_id] ON [dbo].[users];\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableChangeIndexOption(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		  INDEX [ix_users_id] UNIQUE CLUSTERED ([id]) WITH (
		    PAD_INDEX = ON
		  )
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE users (
		  id bigint NOT NULL,
		  name varchar(20),
		  INDEX [ix_users_id] UNIQUE CLUSTERED ([id]) WITH (
		    PAD_INDEX = ON,
		    FILLFACTOR = 10
		  )
		);
		`,
	)

	assertApplyOutput(t, createTable, applyPrefix+"DROP INDEX [ix_users_id] ON [dbo].[users];\n"+"CREATE UNIQUE CLUSTERED INDEX [ix_users_id] ON [dbo].[users] ([id]) WITH (pad_index = ON, fillfactor = 10);\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableForeignKey(t *testing.T) {
	resetTestDatabase()

	createUsers := "CREATE TABLE users (id BIGINT PRIMARY KEY);\n"
	createPosts := stripHeredoc(`
		CREATE TABLE posts (
		  content text,
		  user_id bigint
		);
		`,
	)
	assertApplyOutput(t, createUsers+createPosts, applyPrefix+createUsers+createPosts)
	assertApplyOutput(t, createUsers+createPosts, nothingModified)

	createPosts = stripHeredoc(`
		CREATE TABLE posts (
		  content text,
		  user_id bigint,
		  CONSTRAINT posts_ibfk_1 FOREIGN KEY (user_id) REFERENCES users (id)
		);
		`,
	)
	assertApplyOutput(t, createUsers+createPosts, applyPrefix+"ALTER TABLE [dbo].[posts] ADD CONSTRAINT [posts_ibfk_1] FOREIGN KEY ([user_id]) REFERENCES [users] ([id]);\n")
	assertApplyOutput(t, createUsers+createPosts, nothingModified)

	createPosts = stripHeredoc(`
		CREATE TABLE posts (
		  content text,
		  user_id bigint,
		  CONSTRAINT posts_ibfk_1 FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL ON UPDATE CASCADE
		);
		`,
	)
	assertApplyOutput(t, createUsers+createPosts, applyPrefix+
		"ALTER TABLE [dbo].[posts] DROP CONSTRAINT [posts_ibfk_1];\n"+
		"ALTER TABLE [dbo].[posts] ADD CONSTRAINT [posts_ibfk_1] FOREIGN KEY ([user_id]) REFERENCES [users] ([id]) ON DELETE SET NULL ON UPDATE CASCADE;\n",
	)
	assertApplyOutput(t, createUsers+createPosts, nothingModified)

	createPosts = stripHeredoc(`
		CREATE TABLE posts (
		  content text,
		  user_id bigint
		);
		`,
	)
	assertApplyOutput(t, createUsers+createPosts, applyPrefix+"ALTER TABLE [dbo].[posts] DROP CONSTRAINT [posts_ibfk_1];\n")
	assertApplyOutput(t, createUsers+createPosts, nothingModified)
}

func TestMssqldefCreateTableWithCheck(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE a (
		  a_id INTEGER PRIMARY KEY CONSTRAINT [a_a_id_check] CHECK ([a_id]>(0)),
		  my_text TEXT NOT NULL
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE a (
		  a_id INTEGER PRIMARY KEY CONSTRAINT [a_a_id_check] CHECK ([a_id]>(1)),
		  my_text TEXT NOT NULL
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+
		"ALTER TABLE [dbo].[a] DROP CONSTRAINT a_a_id_check;\n"+
		"ALTER TABLE [dbo].[a] ADD CONSTRAINT a_a_id_check CHECK (a_id > (1));\n")
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE a (
		  a_id INTEGER PRIMARY KEY,
		  my_text TEXT NOT NULL
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+
		"ALTER TABLE [dbo].[a] DROP CONSTRAINT a_a_id_check;\n")
	assertApplyOutput(t, createTable, nothingModified)
}

func TestMssqldefCreateTableWithCheckWithoutName(t *testing.T) {
	resetTestDatabase()

	createTable := stripHeredoc(`
		CREATE TABLE a (
		  a_id INTEGER PRIMARY KEY CHECK ([a_id]>(0)),
		  my_text TEXT NOT NULL
		);
		`,
	)
	assertApplyOutput(t, createTable, applyPrefix+createTable)
	assertApplyOutput(t, createTable, nothingModified)

	createTable = stripHeredoc(`
		CREATE TABLE a (
		  a_id INTEGER PRIMARY KEY CHECK ([a_id]>(1)),
		  my_text TEXT NOT NULL
		);
		`,
	)

	// extract name of check constraint from sql server
	out, err := execute("sqlcmd", "-Usa", "-PPassw0rd", "-dmssqldef_test", "-h", "-1", "-Q", stripHeredoc(`
		SELECT name FROM sys.check_constraints cc WHERE cc.parent_object_id = OBJECT_ID('dbo.a', 'U');
		`,
	))
	if err != nil {
		t.Error("failed to extract check constraint name")
	}
	checkConstraintName := strings.Replace((strings.Split(out, "\n")[0]), " ", "", -1)
	dropConstraint := fmt.Sprintf("ALTER TABLE [dbo].[a] DROP CONSTRAINT %s;\n", checkConstraintName)

	assertApplyOutput(t, createTable, applyPrefix+
		dropConstraint+"ALTER TABLE [dbo].[a] ADD CONSTRAINT a_a_id_check CHECK (a_id > (1));\n")
	assertApplyOutput(t, createTable, nothingModified)
}

//
// ----------------------- following tests are for CLI -----------------------
//

func TestMssqldefDryRun(t *testing.T) {
	resetTestDatabase()
	writeFile("schema.sql", stripHeredoc(`
		CREATE TABLE users (
		  id integer NOT NULL PRIMARY KEY,
		  age integer
		);`,
	))

	dryRun := assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--dry-run", "--file", "schema.sql")
	apply := assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--file", "schema.sql")
	assertEquals(t, dryRun, strings.Replace(apply, "Apply", "dry run", 1))
}

func TestMssqldefSkipDrop(t *testing.T) {
	resetTestDatabase()
	mustExecute("sqlcmd", "-Usa", "-PPassw0rd", "-dmssqldef_test", "-Q", stripHeredoc(`
		CREATE TABLE users (
		    id integer NOT NULL PRIMARY KEY,
		    age integer
		);`,
	))

	writeFile("schema.sql", "")

	skipDrop := assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--skip-drop", "--file", "schema.sql")
	apply := assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--file", "schema.sql")
	assertEquals(t, skipDrop, strings.Replace(apply, "DROP", "-- Skipped: DROP", 1))
}

func TestMssqldefExport(t *testing.T) {
	resetTestDatabase()
	out := assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--export")
	assertEquals(t, out, "-- No table exists --\n")

	mustExecute("sqlcmd", "-Usa", "-PPassw0rd", "-dmssqldef_test", "-Q", stripHeredoc(`
		CREATE TABLE dbo.users (
		    id int NOT NULL,
		    age int
		);
		`,
	))
	out = assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--export")
	assertEquals(t, out, stripHeredoc(`
		CREATE TABLE dbo.users (
		    id int NOT NULL,
		    age int
		);
		`,
	))
}

func TestMssqldefHelp(t *testing.T) {
	_, err := execute("mssqldef", "--help")
	if err != nil {
		t.Errorf("failed to run --help: %s", err)
	}

	out, err := execute("mssqldef")
	if err == nil {
		t.Errorf("no database must be error, but successfully got: %s", out)
	}
}

func TestMain(m *testing.M) {
	resetTestDatabase()
	mustExecute("go", "build")
	status := m.Run()
	os.Exit(status)
}

func assertApply(t *testing.T, schema string) {
	t.Helper()
	writeFile("schema.sql", schema)
	assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--file", "schema.sql")
}

func assertApplyOutput(t *testing.T, schema string, expected string) {
	t.Helper()
	writeFile("schema.sql", schema)
	actual := assertedExecute(t, "mssqldef", "-Usa", "-PPassw0rd", "mssqldef_test", "--file", "schema.sql")
	assertEquals(t, actual, expected)
}

func mustExecute(command string, args ...string) {
	out, err := execute(command, args...)
	if err != nil {
		log.Printf("failed to execute '%s %s': `%s`", command, strings.Join(args, " "), out)
		log.Fatal(err)
	}
}

func assertedExecute(t *testing.T, command string, args ...string) string {
	t.Helper()
	out, err := execute(command, args...)
	if err != nil {
		t.Errorf("failed to execute '%s %s' (error: '%s'): `%s`", command, strings.Join(args, " "), err, out)
	}
	return out
}

func assertEquals(t *testing.T, actual string, expected string) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected '%s' but got '%s'", expected, actual)
	}
}

func execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func resetTestDatabase() {
	mustExecute("sqlcmd", "-Usa", "-PPassw0rd", "-Q", "DROP DATABASE IF EXISTS mssqldef_test;")
	mustExecute("sqlcmd", "-Usa", "-PPassw0rd", "-Q", "CREATE DATABASE mssqldef_test;")
}

func writeFile(path string, content string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Write(([]byte)(content))
}

func stripHeredoc(heredoc string) string {
	heredoc = strings.TrimPrefix(heredoc, "\n")
	re := regexp.MustCompilePOSIX("^\t*")
	return re.ReplaceAllLiteralString(heredoc, "")
}
