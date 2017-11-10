package crdb

import (
	"database/sql"

	"github.com/lib/pq"
	manager "github.com/ory/ladon/manager/sql"
	migrate "github.com/rubenv/sql-migrate"
	gorp "gopkg.in/gorp.v1"
)

func init() {
	if !driverExists(sql.Drivers(), "cockroachdb") {
		sql.Register("cockroachdb", &pq.Driver{})
	}
	migrate.MigrationDialects["cockroachdb"] = gorp.PostgresDialect{}
	manager.Migrations["cockroachdb"] = manager.Statements{
		Migrations: &migrate.MemoryMigrationSource{
			Migrations: []*migrate.Migration{
				{
					Id: "1",
					Up: []string{
						`CREATE TABLE IF NOT EXISTS ladon_policy (
							id           varchar(255) NOT NULL UNIQUE PRIMARY KEY,
							description  text NOT NULL,
							effect       text NOT NULL CHECK (effect='allow' OR effect='deny'),
							conditions   text NOT NULL
						)`,
						`CREATE TABLE IF NOT EXISTS ladon_subject (
							id         varchar(64) NOT NULL UNIQUE PRIMARY KEY,
							has_regex  bool NOT NULL,
							compiled   varchar(511) NOT NULL UNIQUE,
							template   varchar(511) NOT NULL UNIQUE
						)`,
						`CREATE TABLE IF NOT EXISTS ladon_action (
							id         varchar(64) NOT NULL UNIQUE PRIMARY KEY,
							has_regex  bool NOT NULL,
							compiled   varchar(511) NOT NULL UNIQUE,
							template   varchar(511) NOT NULL UNIQUE
						)`,
						`CREATE TABLE IF NOT EXISTS ladon_resource (
							id         varchar(64) NOT NULL UNIQUE PRIMARY KEY,
							has_regex  bool NOT NULL,
							compiled   varchar(511) NOT NULL UNIQUE,
							template   varchar(511) NOT NULL UNIQUE
						)`,
						`CREATE TABLE IF NOT EXISTS ladon_policy_subject_rel (
							policy   varchar(255) NOT NULL,
							subject  varchar(64) NOT NULL,
							PRIMARY KEY (policy, subject),
							FOREIGN KEY (policy) REFERENCES ladon_policy(id),
							FOREIGN KEY (subject) REFERENCES ladon_subject(id)
						)`,
						`CREATE TABLE IF NOT EXISTS ladon_policy_action_rel (
							policy   varchar(255) NOT NULL,
							action  varchar(64) NOT NULL,
							PRIMARY KEY (policy, action),
							FOREIGN KEY (policy) REFERENCES ladon_policy(id),
							FOREIGN KEY (action) REFERENCES ladon_action(id)
						)`,
						`CREATE TABLE IF NOT EXISTS ladon_policy_resource_rel (
							policy    varchar(255) NOT NULL,
							resource  varchar(64) NOT NULL,
							PRIMARY KEY (policy, resource),
							FOREIGN KEY (policy) REFERENCES ladon_policy(id),
							FOREIGN KEY (resource) REFERENCES ladon_resource(id)
						)`,
						`CREATE INDEX ladon_subject_compiled_idx ON ladon_subject (compiled)`,
						`CREATE INDEX ladon_permission_compiled_idx ON ladon_action (compiled)`,
						`CREATE INDEX ladon_resource_compiled_idx ON ladon_resource (compiled)`,
					},
					Down: []string{
						`DROP TABLE ladon_policy`,
						`DROP TABLE ladon_subject`,
						`DROP TABLE ladon_action`,
						`DROP TABLE ladon_resource`,
						`DROP TABLE ladon_policy_subject_rel`,
						`DROP TABLE ladon_policy_action_rel`,
						`DROP TABLE ladon_policy_resource_rel`,
						`DROP INDEX ladon_subject_compiled_idx`,
						`DROP INDEX ladon_permission_compiled_idx`,
						`DROP INDEX ladon_resource_compiled_idx`,
					},
				},
			},
		},
		QueryInsertPolicy:             `INSERT INTO ladon_policy(id, description, effect, conditions) VALUES($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
		QueryInsertPolicyActions:      `INSERT INTO ladon_action (id, template, compiled, has_regex) VALUES($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
		QueryInsertPolicyActionsRel:   `INSERT INTO ladon_policy_action_rel (policy, action) VALUES($1, $2) ON CONFLICT (policy, action) DO NOTHING`,
		QueryInsertPolicyResources:    `INSERT INTO ladon_resource (id, template, compiled, has_regex) VALUES($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
		QueryInsertPolicyResourcesRel: `INSERT INTO ladon_policy_resource_rel (policy, resource) VALUES($1, $2) ON CONFLICT (policy, resource) DO NOTHING`,
		QueryInsertPolicySubjects:     `INSERT INTO ladon_subject (id, template, compiled, has_regex) VALUES($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
		QueryInsertPolicySubjectsRel:  `INSERT INTO ladon_policy_subject_rel (policy, subject) VALUES($1, $2) ON CONFLICT (policy, subject) DO NOTHING`,
		QueryRequestCandidates: `
		SELECT
			p.id,
			p.effect,
			p.conditions,
			p.description,
			subject.template AS subject,
			resource.template AS resource,
			action.template AS action
		FROM
			ladon_policy AS p
			INNER JOIN ladon_policy_subject_rel AS rs ON rs.policy = p.id
			LEFT JOIN ladon_policy_action_rel AS ra ON ra.policy = p.id
			LEFT JOIN ladon_policy_resource_rel AS rr ON rr.policy = p.id
			INNER JOIN ladon_subject AS subject ON rs.subject = subject.id
			LEFT JOIN ladon_action AS action ON ra.action = action.id
			LEFT JOIN ladon_resource AS resource ON rr.resource = resource.id
		WHERE
			(subject.has_regex IS NOT TRUE AND subject.template = $1)
			OR
			(subject.has_regex IS TRUE AND $2 ~ subject.compiled)`,
	}
}

func driverExists(drivers []string, driver string) bool {
	for _, a := range drivers {
		if a == driver {
			return true
		}
	}
	return false
}
