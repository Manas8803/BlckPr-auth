package configs

import (
	"context"
	"log"
	"net/smtp"
	"os"

	"github.com/jackc/pgx/v5"
)

var CONN *pgx.Conn

const connectMsg string = "\n---------------------------------------------------------------------------------------------\nConnected to DB\n---------------------------------------------------------------------------------------------"

func ConnectDB() *pgx.Conn {
	ctx := context.Background()
	uri := os.Getenv("SQLURI")
	conn, err := pgx.Connect(ctx, uri)
	if err != nil {
		log.Println(err)
		NotifyAdmin(err)
		return nil
	}
	CONN = conn

	log.Println(connectMsg)
	return conn
}
func NotifyAdmin(err error) error {
	auth := smtp.PlainAuth("", os.Getenv("EMAIL"), os.Getenv("PASSWORD"), "smtp.gmail.com")

	email := os.Getenv("ADMIN")
	to := []string{email}

	message := []byte(
		"To:" + email + "\r\n" +
			"Subject: Error\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"utf-8\"\r\n\r\n" +
			"<html>" +
			"<head>" +
			"<title>Error in deployment</title>" +
			"</head>" +
			"<body style=\"font-family: Arial, sans-serif;\">" +
			"<div style=\"padding: 20px;\">" +
			"<h1 style=\"color: #333;\">An error occured just now!!!</h1>" +
			"<p style=\"font-size: 16px;\">ERROR : <strong>" + err.Error() + "</strong></p>" +
			"</div>" +
			"</body>" +
			"</html>")

	sendEmailErr := smtp.SendMail("smtp.gmail.com:587", auth, os.Getenv("EMAIL"), to, message)

	log.Fatal(sendEmailErr)
	return sendEmailErr
}

func CloseDB() {
	if err := CONN.Close(context.Background()); err != nil {
		log.Fatalf("Error closing database connection: %v", err)
	}
}