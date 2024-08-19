// main.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}
)

// Task represents a task in the application
type Task struct {
	Id            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Completed     bool   `json:"completed"`
	UserID        string `json:"user_id"`
	StartTime     string `json:"start_time"`
	CompletedTime string `json:"completed_time"`
}

func main() {
	// Initialize Gin router
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	router.GET("/auth/login", func(ctx *gin.Context) {
		url := oauthConfig.AuthCodeURL("state")
		ctx.Redirect(http.StatusFound, url)
	})

	router.GET("/auth/callback", func(ctx *gin.Context) {
		code := ctx.Query("code")
		token, err := oauthConfig.Exchange(ctx, code)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Error exchange code for token")
			return
		}

		ctx.String(http.StatusOK, "Authention successful! Access token: %s", token.AccessToken)
		ctx.Redirect(http.StatusFound, "/")
	})

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/", func(ctx *gin.Context) {

		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/public", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "This is a public resource")
	})

	router.POST("/tasks", createTaskHandler)
	router.GET("/tasks", getTasksHandler)
	router.PUT("/update-task/:id", updateTaskHandler)
	router.DELETE("/delete-task/:id", deleteTaskHandler)
	router.PUT("/complete-task/:id", completeTaskHandler)

	// Run the server
	router.Run(":8080")
}

func isAuthenticated() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken := ctx.GetHeader("Authorization")
		if accessToken == "" {

			ctx.Redirect(http.StatusFound, "/auth/login")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func DBConnection() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_DATABASE")

	// Connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return connStr
}

func deleteTaskHandler(ctx *gin.Context) {
	taskId := ctx.Param("id")
	var task Task
	if err := ctx.ShouldBind(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	db, err := sql.Open("postgres", DBConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query to retrieve tasks from the database
	_, err = db.Exec("DELETE FROM tasks WHERE id = $1", taskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	fmt.Println("Data deleted successfully!")
}

func completeTaskHandler(ctx *gin.Context) {
	taskId := ctx.Param("id")
	var task Task
	if err := ctx.ShouldBind(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	// Open a connection to the database
	db, err := sql.Open("postgres", DBConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	currentTime := time.Now().Format("01/02/06 15:04:05")

	// Query to retrieve tasks from the database
	_, err = db.Exec("UPDATE tasks SET completed = $1, completed_time = $2 WHERE id = $3",
		task.Completed, currentTime, taskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	fmt.Println("Data updated successfully!")
}

func updateTaskHandler(ctx *gin.Context) {
	taskId := ctx.Param("id")
	var task Task
	if err := ctx.ShouldBind(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	// Open a connection to the database
	db, err := sql.Open("postgres", DBConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query to retrieve tasks from the database
	_, err = db.Exec("UPDATE tasks SET title =$1, description = $2 WHERE id = $3",
		task.Title, task.Description, taskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	fmt.Println("Data updated successfully!")
}

func createTaskHandler(ctx *gin.Context) {
	var task Task

	if err := ctx.ShouldBind(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	ctx.Status(http.StatusCreated)

	if err := writeDataToSQL(&task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}
}

func getTasksHandler(ctx *gin.Context) {

	// Open a connection to the database
	db, err := sql.Open("postgres", DBConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query to retrieve tasks from the database
	rows, err := db.Query("SELECT id, title, description, user_id, start_time, completed_time, completed FROM tasks ORDER BY start_time DESC")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			`error`: err.Error()})
		return
	}
	defer rows.Close()

	// Iterate through the rows and populate the taskList
	var taskList []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.Id, &task.Title, &task.Description, &task.UserID, &task.StartTime, &task.CompletedTime, &task.Completed)
		if err != nil {
			log.Fatal(err)
		}
		taskList = append(taskList, task)
	}
	if err := rows.Err(); err != nil {

		log.Fatal(err)
	}

	// Return the taskList as JSON response
	ctx.JSON(http.StatusOK, taskList)
}

func writeDataToSQL(t *Task) error {

	tableName := os.Getenv("DB_TABLE")
	// Open a connection to the database
	db, err := sql.Open("postgres", DBConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Insert data into the table
	insertStmt := fmt.Sprintf("INSERT INTO %s(title, description, completed, user_id, start_time, completed_time) VALUES ($1, $2, $3, $4, $5, $6)", tableName)
	_, err = db.Exec(insertStmt, t.Title, t.Description, t.Completed, t.UserID, t.StartTime, t.CompletedTime)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Data inserted successfully!")

	return nil
}
