package db

import (
	"database/sql"
	"os"
	"path/filepath"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string)(*Store,error){
	if dir:=filepath.Dir(path);dir!="."{
		if err:=os.MkdirAll(dir,0755);err!=nil{return nil,err}
	}
	sqlDB,err:=sql.Open("sqlite",path);if err!=nil{return nil,err}
	s:=&Store{sqlDB}
	if err:=s.migrate();err!=nil{sqlDB.Close();return nil,err}
	return s,nil
}
func(s *Store)Close()error{return s.db.Close()}

func(s *Store)migrate()error{
	_,err:=s.db.Exec(`
PRAGMA journal_mode=WAL;
CREATE TABLE IF NOT EXISTS runs(
 id TEXT PRIMARY KEY,status TEXT NOT NULL,input TEXT NOT NULL,
 output TEXT NOT NULL DEFAULT '',error TEXT NOT NULL DEFAULT '',
 created_at INTEGER NOT NULL,updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS steps(
 id TEXT PRIMARY KEY,run_id TEXT NOT NULL,step_no INTEGER NOT NULL,
 kind TEXT NOT NULL,status TEXT NOT NULL,input TEXT NOT NULL DEFAULT '',
 output TEXT NOT NULL DEFAULT '',error TEXT NOT NULL DEFAULT '',
 created_at INTEGER NOT NULL,updated_at INTEGER NOT NULL,
 FOREIGN KEY(run_id) REFERENCES runs(id)
);
CREATE INDEX IF NOT EXISTS idx_steps_run ON steps(run_id,step_no);
CREATE TABLE IF NOT EXISTS tool_calls(
 id TEXT PRIMARY KEY,run_id TEXT NOT NULL,step_id TEXT NOT NULL,
 tool_name TEXT NOT NULL,arguments TEXT NOT NULL,status TEXT NOT NULL,
 result TEXT NOT NULL DEFAULT '',error TEXT NOT NULL DEFAULT '',
 created_at INTEGER NOT NULL,updated_at INTEGER NOT NULL,
 FOREIGN KEY(run_id) REFERENCES runs(id),
 FOREIGN KEY(step_id) REFERENCES steps(id)
);
CREATE INDEX IF NOT EXISTS idx_tool_calls_run ON tool_calls(run_id,created_at);
`)
	return err
}
