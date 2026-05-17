package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // РґСЂР°Р№РІРµСЂ СЂРµРіРёСЃС‚СЂРёСЂСѓРµС‚СЃСЏ С‡РµСЂРµР· init()

	"github.com/omnikk/pz6/services/tasks/internal/service"
)

type PostgresRepo struct {
	db *sql.DB
}

// NewPostgres РѕС‚РєСЂС‹РІР°РµС‚ СЃРѕРµРґРёРЅРµРЅРёРµ Рё РїСЂРѕРІРµСЂСЏРµС‚ РµРіРѕ РїРёРЅРіРѕРј.
// РќР° РІС…РѕРґ вЂ” РіРѕС‚РѕРІС‹Р№ DSN (С„РѕСЂРјР°С‚ СЃРј. РІ main.go).
func NewPostgres(dsn string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// РґРѕ РїРµСЂРІРѕРіРѕ СЂРµР°Р»СЊРЅРѕРіРѕ Р·Р°РїСЂРѕСЃР° СЃРѕРµРґРёРЅРµРЅРёРµ РЅРµ РїСЂРѕРІРµСЂСЏРµС‚СЃСЏ вЂ” РїРёРЅРіСѓРµРј СЃР°РјРё
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	return &PostgresRepo{db: db}, nil
}

func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// ----- CRUD -----

func (r *PostgresRepo) Create(t service.Task) (service.Task, error) {
	var dueDate any
	if t.DueDate == "" {
		dueDate = nil
	} else {
		dueDate = t.DueDate
	}

	const q = `
		INSERT INTO tasks (id, title, description, due_date, done)
		VALUES ($1, $2, $3, $4::date, $5)
		RETURNING id, title, description, COALESCE(due_date::text, ''), done`

	var out service.Task
	err := r.db.QueryRow(q, t.ID, t.Title, t.Description, dueDate, t.Done).
		Scan(&out.ID, &out.Title, &out.Description, &out.DueDate, &out.Done)
	if err != nil {
		return service.Task{}, fmt.Errorf("insert task: %w", err)
	}
	return out, nil
}

func (r *PostgresRepo) List() ([]service.Task, error) {
	const q = `
		SELECT id, title, description, COALESCE(due_date::text, ''), done
		FROM tasks
		ORDER BY created_at`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *PostgresRepo) Get(id string) (service.Task, bool, error) {
	const q = `
		SELECT id, title, description, COALESCE(due_date::text, ''), done
		FROM tasks WHERE id = $1`

	var t service.Task
	err := r.db.QueryRow(q, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done)
	if err == sql.ErrNoRows {
		return service.Task{}, false, nil
	}
	if err != nil {
		return service.Task{}, false, fmt.Errorf("get task: %w", err)
	}
	return t, true, nil
}

func (r *PostgresRepo) Update(id string, title *string, done *bool) (service.Task, bool, error) {
	// COALESCE($1, title) вЂ” РµСЃР»Рё title == nil, РѕСЃС‚Р°РІР»СЏРµРј СЃС‚Р°СЂРѕРµ Р·РЅР°С‡РµРЅРёРµ.
	// Р­С‚Рѕ РїРѕР·РІРѕР»СЏРµС‚ РѕР±РЅРѕРІР»СЏС‚СЊ РІС‹Р±РѕСЂРѕС‡РЅРѕ РѕРґРЅРёРј Р·Р°РїСЂРѕСЃРѕРј.
	const q = `
		UPDATE tasks
		SET title = COALESCE($2, title),
		    done  = COALESCE($3, done),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, description, COALESCE(due_date::text, ''), done`

	var t service.Task
	err := r.db.QueryRow(q, id, title, done).
		Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done)
	if err == sql.ErrNoRows {
		return service.Task{}, false, nil
	}
	if err != nil {
		return service.Task{}, false, fmt.Errorf("update task: %w", err)
	}
	return t, true, nil
}

func (r *PostgresRepo) Delete(id string) (bool, error) {
	res, err := r.db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete task: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ----- Search: Р±РµР·РѕРїР°СЃРЅС‹Р№ Рё СѓСЏР·РІРёРјС‹Р№ -----

// Search вЂ” РїР°СЂР°РјРµС‚СЂРёР·РѕРІР°РЅРЅС‹Р№ Р·Р°РїСЂРѕСЃ. Р’РІРѕРґ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ РќР• РёРЅС‚РµСЂРїСЂРµС‚РёСЂСѓРµС‚СЃСЏ
// РєР°Рє С‡Р°СЃС‚СЊ SQL: РґСЂР°Р№РІРµСЂ РїРµСЂРµРґР°С‘С‚ РµРіРѕ РєР°Рє Р·РЅР°С‡РµРЅРёРµ РїР°СЂР°РјРµС‚СЂР°.
func (r *PostgresRepo) Search(title string) ([]service.Task, error) {
	const q = `
		SELECT id, title, description, COALESCE(due_date::text, ''), done
		FROM tasks WHERE title = $1`

	rows, err := r.db.Query(q, title)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

// SearchVulnerable вЂ” РќРђРњР•Р Р•РќРќРћ СѓСЏР·РІРёРјР°СЏ СЂРµР°Р»РёР·Р°С†РёСЏ РґР»СЏ РґРµРјРѕРЅСЃС‚СЂР°С†РёРё.
// РљРѕРЅРєР°С‚РµРЅР°С†РёСЏ РІРІРѕРґР° РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ СЃ SQL = РєР»Р°СЃСЃРёС‡РµСЃРєР°СЏ SQL-РёРЅСЉРµРєС†РёСЏ.
// РџРµР№Р»РѕР°Рґ `' OR '1'='1` РїСЂРµРІСЂР°С‚РёС‚ Р·Р°РїСЂРѕСЃ РІ:
//
//	SELECT ... WHERE title = '' OR '1'='1'
//
// Рё РІРµСЂРЅС‘С‚ Р’РЎР• СЃС‚СЂРѕРєРё.
func (r *PostgresRepo) SearchVulnerable(title string) ([]service.Task, error) {
	q := "SELECT id, title, description, COALESCE(due_date::text, ''), done " +
		"FROM tasks WHERE title = '" + title + "'"

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("vulnerable search: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

// scanTasks вЂ” РѕР±С‰РёР№ РїРѕРјРѕС‰РЅРёРє РґР»СЏ СЃС‚СЂРѕРє-СЂРµР·СѓР»СЊС‚Р°С‚РѕРІ.
func scanTasks(rows *sql.Rows) ([]service.Task, error) {
	result := make([]service.Task, 0)
	for rows.Next() {
		var t service.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return result, nil
}
