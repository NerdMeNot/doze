package documentdb

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nerdmenot/doze/internal/engine"
)

// provisioned reports whether the private Postgres cluster has been initialized.
func provisioned(dataDir string) bool {
	_, err := os.Stat(filepath.Join(pgDataDir(dataDir), "PG_VERSION"))
	return err == nil
}

// provision initializes the private Postgres cluster if needed and (re)writes
// the DocumentDB-required configuration. Idempotent.
func provision(ctx context.Context, inst engine.Instance, tc engine.Toolchain) error {
	pgData := pgDataDir(inst.DataDir)
	if _, err := os.Stat(filepath.Join(pgData, "PG_VERSION")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := initdb(ctx, inst, tc, pgData); err != nil {
			return err
		}
	}
	if err := writeConf(pgData); err != nil {
		return err
	}
	return writeHBA(pgData)
}

func initdb(ctx context.Context, inst engine.Instance, tc engine.Toolchain, pgData string) error {
	if err := os.MkdirAll(filepath.Dir(pgData), 0o700); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, tc.Path("initdb"),
		"-D", pgData,
		"-U", "postgres",
		"-A", "trust",
		"-E", "UTF8",
		"--no-sync",
	)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("initdb for documentdb %q failed: %w\n%s", inst.Name, err, out.String())
	}
	return nil
}

// writeConf writes the DocumentDB-required settings into doze.conf, included by
// the initdb-generated postgresql.conf (so we never clobber that file). These
// are mandatory for the extension to load and operate:
//   - shared_preload_libraries: pg_cron + the two documentdb shared libs must be
//     loaded at server start.
//   - cron.database_name: pg_cron runs its scheduler against this database, which
//     is also where we create the extension.
//   - listen_addresses=127.0.0.1: the extension self-connects over loopback TCP.
//
// The fsync/durability settings mirror doze's light dev profile.
func writeConf(pgData string) error {
	const conf = `# Managed by doze — do not edit. Regenerated on every boot.
listen_addresses = '127.0.0.1'
shared_preload_libraries = 'pg_cron,pg_documentdb_core,pg_documentdb'
cron.database_name = 'postgres'
fsync = off
synchronous_commit = off
full_page_writes = off
`
	if err := os.WriteFile(filepath.Join(pgData, "doze.conf"), []byte(conf), 0o600); err != nil {
		return err
	}
	return ensureInclude(pgData)
}

func ensureInclude(pgData string) error {
	mainConf := filepath.Join(pgData, "postgresql.conf")
	data, err := os.ReadFile(mainConf)
	if err != nil {
		return err
	}
	const directive = "include = 'doze.conf'"
	if bytes.Contains(data, []byte(directive)) {
		return nil
	}
	f, err := os.OpenFile(mainConf, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n# Added by doze\n%s\n", directive)
	return err
}

// writeHBA permits local socket trust (FerretDB and our psql) and loopback TCP
// trust (the extension's self-connection). Local dev only — never exposed.
func writeHBA(pgData string) error {
	const hba = `# Managed by doze — local + loopback trust only.
# TYPE  DATABASE  USER  ADDRESS        METHOD
local   all       all                  trust
host    all       all   127.0.0.1/32   trust
host    all       all   ::1/128        trust
`
	return os.WriteFile(filepath.Join(pgData, "pg_hba.conf"), []byte(hba), 0o600)
}
