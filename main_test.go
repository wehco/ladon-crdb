package crdb_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ory/dockertest"
	manager "github.com/ory/ladon/manager/sql"

	_ "github.com/wehco/ladon-crdb"
)

var (
	crdb      *sqlx.DB
	resources []*dockertest.Resource
	pool      *dockertest.Pool
)

func init() {
	sql.Register("cockroachdb", &pq.Driver{})
}

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(1)
	crdb = ConnectToCockroachDB(&wg)
	wg.Wait()

	s := m.Run()
	KillAll()
	os.Exit(s)
}

func KillAll() {
	for _, resource := range resources {
		pool.Purge(resource)
	}
	resources = []*dockertest.Resource{}
}

func ConnectToCockroachDB(wg *sync.WaitGroup) *sqlx.DB {
	var (
		db  *sqlx.DB
		err error
	)

	defer wg.Done()

	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "cockroachdb/cockroach",
		Tag:        "v1.1.1",
		Cmd:        []string{"start", "--insecure"},
	})
	if err != nil {
		log.Fatalf("Could not start CockroachDB: %v", err)
	}

	if err = pool.Retry(func() error {
		var err error
		db, err = sqlx.Open(
			"cockroachdb",
			fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=disable",
				"root",
				"",
				"localhost",
				resource.GetPort("26257/tcp"),
				"ladon",
			),
		)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}
	resources = append(resources, resource)
	return db
}

func TestAll(t *testing.T) {
	if _, err := crdb.Exec("CREATE DATABASE IF NOT EXISTS ladon;"); err != nil {
		log.Fatalf("Could not create database: %v", err)
	}
	s := manager.NewSQLManager(crdb, nil)
	t.Run("Migrations", func(t *testing.T) {
		_, err := s.CreateSchemas("", "ladon_migrations")
		if err != nil {
			t.Fatalf("CreateSchemas error: %v", err)
		}
	})
}
