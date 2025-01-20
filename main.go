package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"
	"github.com/labstack/echo/v4"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	initDB()
	defer db.Close()

    e := echo.New()

	e.POST("/todo", createTodo)
	e.GET("/todo", getTodo)
	e.PUT("/todo/:id", putTodo)
	e.DELETE("/todo/:id", deleteTodo)

	e.Start(":2004")
}

//データベースの設計

type todo struct {
	ID         int        `json:"id"`
	Title      string     `json:"title"`
	Regtime    time.Time  `json:"regtime"`
	Renewtime  time.Time  `json:"renewtime"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./todo.db")
	if err != nil {
		log.Fatal(err)
	} 

//テーブル作成
    sqlStmt := `
	CREATE TABLE IF NOT EXISTS todos (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		regtime DATETIME DEFAULT CURRENT_TIMESTAMP,
	renewtime DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("sqltable error: %v", err)
	}
	log.Println("OKOK")
}

//POSTTODO作成API
type createTodoRequest struct {
    Title string `json:"title"`
}

//type createTodoResponse struct {
// ID int `json:"id"`
//}

func createTodo (c echo.Context) error {
	req := new(createTodoRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message":"503"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message": "503"})
	}

	sqlStmt := "INSERT INTO todos (title) VALUES (?)"
	_, err := db.Exec(sqlStmt, req.Title)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message":"503"})
	}

	return c.NoContent((http.StatusNoContent))
	//id, _ := result.LastInsertId()
	//return c.JSON(http.StatusCreated, createTodoResponse{ID: int(id)})


	//return c.NoContent(http.StatusNoContent)

}




//GETTODO一覧取得API

type getTodoResponse struct {
	Todos []todo `json:"todos"`
}

func getTodo (c echo.Context) error {
	title := c.QueryParam("title")
	query := `SELECT id, title, renewtime FROM todos`
    
	var err error
//chatGPT使用↓↓
    var rows *sql.Rows
if title != "" {
	query += " WHERE title LIKE ?"
	rows, err = db.Query(query, "%"+title+"%")
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message": "Failed to retrieve todos"})
	}
} else {
	rows, err = db.Query(query)

	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message": "Failed to retrieve todos"})
	}
}
//↑
    defer rows.Close()

    var todos []todo
	for rows.Next() {
		var t todo
		err := rows.Scan(&t.ID, &t.Title, &t.Renewtime)
		if err != nil {
			log.Println(err)
			continue
		}
		todos = append(todos, t)
	}
	if len(todos) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found"})
	}
	return c.JSON(http.StatusOK, getTodoResponse{Todos: todos})
}



//PUTTODO更新API

func putTodo (c echo.Context) error {
	req := new(createTodoRequest)
	//title := c.Bind("title")
	id := c.Param("id")

    if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input format"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "503"})
	}

	//if Title == "" {
		//return c.JSON(http.StatusBadRequest,map[string]string{"message": "Not found"})
	//}

	stmt, err := db.Prepare("UPDATE todos SET title = ?, renewtime = CURRENT_TIMESTAMP WHERE id = ?")
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable,map[string]string{"message": "503"})
	}
	result, err := stmt.Exec(req.Title,id)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable,map[string]string{"message": "503"})
	}
	//chatGPT↓
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	//↑
	return c.NoContent(http.StatusNoContent)
}	



//DELETETODO削除API

func deleteTodo (c echo.Context) error {
	id := c.Param("id")

	stmt, err := db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message": "503"})
	}

	result, err := stmt.Exec(id)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message": "error"})
	}

	//chatGPT↓
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.NoContent(http.StatusNotFound)
	}

	return c.NoContent(http.StatusNoContent)
}