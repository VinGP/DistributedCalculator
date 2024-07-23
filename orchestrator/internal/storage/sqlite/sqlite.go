package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/entities"
	"github.com/vingp/DistributedCalculator/orchestrator/internal/storage"
)

type Stor interface {
	// Users
	SaveUser(username string, passwordHash string) (entities.User, error)
	GetUser(id int64) (entities.User, error)

	// Expressions
	SaveExpression(expression *entities.Expression) error
	GetExpression(id string) (entities.Expression, error)
	DeleteExpression(expressionId string) error
	UpdateExpression(expression *entities.Expression) error

	// RPN Tokens
	GetRPNTokens(expressionId string) ([]*entities.RPNToken, error)
	SaveRPNTokens(tokens []*entities.RPNToken) error
	DeleteRPNTokens(rpnTokenId int) error
	DeleteExpressionRPNTokens(expressionId int) error

	// Tasks
	GetNewTask() ([]entities.Task, error)
	AllTaskDone(expressionId string) (bool, error)
	GetTasks(expressionId string) ([]entities.Task, error)
	GetTaskById(id string) (entities.Task, error)
	SaveTask(task *entities.Task) error
	UpdateTask(task *entities.Task) error
	DeleteTask(id string) error
}

type Storage struct {
	db *sql.DB
}

func createTables(db *sql.DB) error {
	const op = "internal.storage.sqlite.createTables"

	var (
		createUserTableStmt   *sql.Stmt
		createExpressionsStmt *sql.Stmt
		createRpnTokensStmt   *sql.Stmt
		createTasksStmt       *sql.Stmt
		err                   error
	)

	// Prepare statements for each table
	createUserTableStmt, err = db.Prepare(`
	  CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	  );
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	createExpressionsStmt, err = db.Prepare(`
	  CREATE TABLE IF NOT EXISTS expressions (
		id INTEGER PRIMARY KEY,
		expression TEXT NOT NULL,
		result TEXT,
		user_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at TEXT NOT NULL,
		error TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	  );
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	createRpnTokensStmt, err = db.Prepare(`
  CREATE TABLE IF NOT EXISTS rpn_tokens (
    id INTEGER PRIMARY KEY,
    expression_id INTEGER NOT NULL,
    value TEXT NOT NULL,
    type TEXT NOT NULL,
    position INTEGER NOT NULL,
    FOREIGN KEY(expression_id) REFERENCES expressions(id)
  );
`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	createTasksStmt, err = db.Prepare(`
  CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY,
    expression_id INTEGER NOT NULL,
    arg1 TEXT NOT NULL,
    arg2 TEXT NOT NULL,
    operation TEXT NOT NULL,
    operation_time INTEGER NOT NULL,
    completed BOOLEAN NOT NULL,
    error TEXT,
    sent_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(expression_id) REFERENCES expressions(id)
  );
`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Execute each statement individually
	_, err = createUserTableStmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = createExpressionsStmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = createRpnTokensStmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = createTasksStmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func New(storagePath string) (*Storage, error) {
	const op = "internal.storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(username string, passwordHash string) (int64, error) {
	const op = "internal.storage.sqlite.SaveUser"

	res, err := s.db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUser(id int64) (entities.User, error) {
	const op = "internal.storage.sqlite.GetUser"

	var user entities.User
	err := s.db.QueryRow("SELECT id, username, password_hash FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return entities.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (s *Storage) SaveExpression(expression *entities.Expression) error {
	const op = "internal.storage.sqlite.SaveExpression"

	stmt, err := s.db.Prepare(`
        INSERT INTO expressions (expression, result, user_id, status, created_at, error)
        VALUES (?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(expression.Expr, expression.Result, expression.UserId, expression.Status, expression.CreatAt, expression.Err)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}
	for _, token := range expression.RPN {
		token.ExpressionId = int(id)
	}
	err = s.SaveRPNTokens(expression.RPN)

	if err != nil {
		return fmt.Errorf("%s: failed to save RPN tokens: %w", op, err)
	}

	expression.Id = int(id)

	return nil
}

func (s *Storage) GetExpression(id string) (entities.Expression, error) {
	const op = "internal.storage.sqlite.GetExpression"

	row := s.db.QueryRow("SELECT * FROM expressions WHERE id = ?", id)
	var expression entities.Expression
	err := row.Scan(&expression.Id, &expression.Expr, &expression.Result, &expression.UserId, &expression.Status, &expression.CreatAt, &expression.Err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Expression{}, fmt.Errorf("%s: %w", op, storage.ErrExpressionNotFound)
		}
		return entities.Expression{}, fmt.Errorf("%s: %w", op, err)
	}

	rpnTokens, err := s.GetRPNTokens(expression.Id)
	if err != nil {
		return entities.Expression{}, fmt.Errorf("%s: %w", op, err)
	}

	expression.RPN = rpnTokens

	return expression, nil
}

func (s *Storage) DeleteExpression(expressionId string) error {
	const op = "internal.storage.sqlite.DeleteExpression"

	_, err := s.db.Exec("DELETE FROM expressions WHERE id = ?", expressionId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) UpdateExpression(expression *entities.Expression) error {
	const op = "internal.storage.sqlite.UpdateExpression"

	_, err := s.db.Exec("UPDATE expressions SET expression = ?, result = ?, status = ?, error = ? WHERE id = ?", expression.Expr, expression.Result, expression.Status, expression.Err, expression.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetRPNTokens(expressionId int) ([]*entities.RPNToken, error) {
	const op = "internal.storage.sqlite.GetRPNTokens"

	rows, err := s.db.Query("SELECT id, expression_id, value, type FROM rpn_tokens WHERE expression_id = ?", expressionId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var rpnTokens []*entities.RPNToken
	for rows.Next() {
		var rpnToken entities.RPNToken
		if err := rows.Scan(&rpnToken.Id, &rpnToken.ExpressionId, &rpnToken.Value, &rpnToken.Type); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		rpnTokens = append(rpnTokens, &rpnToken)
	}
	return rpnTokens, nil
}

func (s *Storage) SaveRPNTokens(tokens []*entities.RPNToken) error {
	const op = "internal.storage.sqlite.SaveRPNTokens"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := tx.Prepare("INSERT INTO rpn_tokens (expression_id, value, type, position) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	for _, token := range tokens {
		_, err = stmt.Exec(token.ExpressionId, token.Value, token.Type, token.Pos)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return tx.Commit()
}

func (s *Storage) DeleteExpressionRPNTokens(expressionId int) error {
	const op = "internal.storage.sqlite.DeleteRPNTokens"

	_, err := s.db.Exec("DELETE FROM rpn_tokens WHERE expression_id = ?", expressionId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) DeleteRPNTokens(rpnTokenId int) error {
	const op = "internal.storage.sqlite.DeleteRPNTokens"

	_, err := s.db.Exec("DELETE FROM rpn_tokens WHERE id = ?", rpnTokenId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

//GetNewTask() ([]entities.Task, error)
//AllTaskDone(expressionId string) (bool, error)
//GetTasks(expressionId string) ([]entities.Task, error)
//GetTaskById(id string) (entities.Task, error)
//SaveTask(task *entities.Task) error
//UpdateTask(task *entities.Task) error
//DeleteTask(id string) error

func (s *Storage) GetNewTask() ([]entities.Task, error) {
	const op = "internal.storage.sqlite.GetNewTask"

	rows, err := s.db.Query("SELECT id, expression_id, status FROM tasks WHERE status = 'new'")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var tasks []entities.Task
	for rows.Next() {
		var task entities.Task
		if err := rows.Scan(&task.Id, &task.ExpressionId, &task.Status); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

type Value int

func (v Value) String() string {
	return [...]string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}[v]
}
