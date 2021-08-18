package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type form struct {
	email   string
	subject string
	message string
}

const (
	dbname = "hngi8db"
)

func dbConnection() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return nil, err
	}
	//defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+dbname)
	if err != nil {
		log.Printf("Error %s when creating DB\n", err)
		return nil, err
	}
	no, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows", err)
		return nil, err
	}
	log.Printf("rows affected %d\n", no)

	db.Close()
	db, err = sql.Open("mysql", dsn())
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return nil, err
	}
	//defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Errors %s pinging DB", err)
		return nil, err
	}
	log.Printf("Connected to DB %s successfully\n", dbname)
	return db, nil
}

func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS contact_me2(email text, subject text, message text)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating contact_me table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when getting rows affected", err)
		return err
	}
	log.Printf("Rows affected when creating table: %d", rows)
	return nil
}

func insert(db *sql.DB, f form) error {
	query := "INSERT INTO contact_me2(email, subject, message) VALUES (?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, f.email, f.subject, f.message)
	if err != nil {
		log.Printf("Error %s when inserting row into contact_me table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return err
	}
	log.Printf("%d contact_me created ", rows)
	return nil
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()
	router.Static("/", "./template")

	router.POST("/submit_form", func(c *gin.Context) {
		f := form{
			email:   c.PostForm("email"),
			subject: c.PostForm("subject"),
			message: c.PostForm("message"),
		}

		db, err := dbConnection()

		if err != nil {
			log.Printf("error: %s when opening DB\n", err)
			c.String(http.StatusInternalServerError, fmt.Sprintf("Internal server error1"))
		}
		defer db.Close()

		log.Printf("Successfully connected to database")
		err = createTable(db)
		if err != nil {
			log.Printf("Create product table failed with error %s", err)
			c.String(http.StatusInternalServerError, fmt.Sprintf("Internal server error2"))
			return
		}

		err = insert(db, f)
		if err != nil {
			log.Printf("Insert product failed with error %s", err)
			c.String(http.StatusInternalServerError, fmt.Sprintf("Internal server error3"))
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("%s! Thank you for getting in touch with me", f.email))
	})

	router.Run(":" + port)
}

func dsn() string {
	url := os.Getenv("DATABASE_URL")

	if url == "" {
		log.Fatal("$DATABASE_URL is not set")
		return ""
	}
	return url
}
