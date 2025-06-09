package main

import (
	"database/sql"
	"github.com/JozefPlata/nlp-phone-book/pkg/cmd"
	"github.com/JozefPlata/nlp-phone-book/pkg/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"net/http"
	"os"
	"strings"
)

type Contact struct {
	Name  string
	Phone string
}

type Response struct {
	Error       string    `json:"error"`
	Message     string    `json:"message"`
	Query       string    `json:"query"`
	Contacts    []Contact `json:"contacts"`
	HasContacts bool      `json:"has_contacts"`
}

func main() {
	//_ = os.Remove("./contacts.db")

	// Init sqlite
	db, err := sql.Open("sqlite3", "./data/contacts.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a table
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS contacts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        phone_number TEXT NOT NULL
    );`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Check for API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	// Use these models only
	var models []string
	models = append(models, openai.ChatModelGPT3_5Turbo)
	models = append(models, openai.ChatModelGPT4oMini)
	models = append(models, openai.ChatModelO1Mini)

	// Cache last command for deletion command
	var lastCommand *cmd.Command

	// Server
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Renderer = templ.NewTemplate()

	// --- Home ---
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", map[string]interface{}{
			"models": models,
		})
	})

	// --- Http chatbot ---
	e.POST("/api/query", func(c echo.Context) error {
		input := c.FormValue("queryInput")
		if input == "" {
			return c.NoContent(http.StatusOK)
		}

		query := strings.TrimSpace(input)
		command, err := cmd.UserQuery(query).Parse(&client, models[0]) // hardcoded gpt-3.5-turbo, worked fine in dev
		if err != nil {
			return c.Render(http.StatusOK, "response-element", Response{
				Error: err.Error(),
				Query: "",
			})
		}
		e.Logger.Info("Command: ", command)

		switch command.Action {
		// ------------------------------------------------------------------------------------------------------------|
		case cmd.ActionCreate:
			lastCommand = &cmd.Command{}
			if command.Name != "" && command.Phone != "" {
				_, err = db.Exec(
					`INSERT INTO contacts (name, phone_number) VALUES (?, ?)`,
					command.Name, command.Phone,
				)
				if err != nil {
					e.Logger.Error(err)
					return c.Render(http.StatusOK, "response-element", Response{
						Error: "Failed to create contact",
						Query: query,
					})
				}
				return c.Render(http.StatusOK, "response-element", Response{
					Message: command.Message + command.Phone,
					Query:   query,
				})
			} else {
				return c.Render(http.StatusOK, "response-element", Response{
					Error: "[C] Name or Phone number is empty",
					Query: query,
				})
			}

		// ------------------------------------------------------------------------------------------------------------|
		case cmd.ActionRead:
			lastCommand = &cmd.Command{}
			if command.Name == cmd.NameAll {
				e.Logger.Info("Read all contacts")
				rows, err := db.Query("SELECT * FROM contacts")
				if err != nil {
					return c.Render(http.StatusOK, "response-element", Response{
						Error: "Something went wrong...",
						Query: query,
					})
				}
				defer rows.Close()

				var contacts []Contact
				for rows.Next() {
					var id int
					var name, phoneNumber string
					err := rows.Scan(&id, &name, &phoneNumber)
					if err != nil {
						continue
					}
					contacts = append(contacts, Contact{name, phoneNumber})
				}
				e.Logger.Info("Contacts:", contacts)
				return c.Render(http.StatusOK, "response-element", Response{
					Contacts:    contacts,
					HasContacts: true,
					Query:       query,
				})

			} else if command.Name != "" || command.Phone != "" {
				row := db.QueryRow(
					`SELECT name, phone_number FROM contacts WHERE name = ? OR phone_number = ?`,
					command.Name, command.Phone,
				)
				var name, phoneNumber string
				err := row.Scan(&name, &phoneNumber)
				if err != nil {
					e.Logger.Error(err)
					return c.Render(http.StatusOK, "response-element", Response{
						Error: "Not found!",
						Query: query,
					})
				}
				if command.Name != "" {
					return c.Render(http.StatusOK, "response-element", Response{
						Message: command.Message + " " + phoneNumber,
						Query:   query,
					})
				} else {
					return c.Render(http.StatusOK, "response-element", Response{
						Message: command.Message + " " + name,
						Query:   query,
					})
				}

			} else {
				return c.Render(http.StatusOK, "response-element", Response{
					Error: "[R] Failed to find contact",
					Query: query,
				})
			}

		// ------------------------------------------------------------------------------------------------------------|
		case cmd.ActionUpdate:
			lastCommand = &cmd.Command{}
			if command.Name != "" && command.Phone != "" {
				_, err := db.Exec(
					`UPDATE contacts SET phone_number = ? WHERE name = ?`,
					command.Phone, command.Name,
				)
				if err != nil {
					e.Logger.Error(err)
					return c.Render(http.StatusOK, "response-element", Response{
						Error: "Failed to update contact",
						Query: query,
					})
				}
				return c.Render(http.StatusOK, "response-element", Response{
					Message: command.Message + " " + command.Phone,
					Query:   query,
				})
			} else {
				return c.Render(http.StatusOK, "response-element", Response{
					Error: "[U] Name or Phone number is empty",
					Query: query,
				})
			}

		// ------------------------------------------------------------------------------------------------------------|
		case cmd.ActionDelete:
			if command.Name != "" || command.Phone != "" {
				lastCommand = command
				return c.Render(http.StatusOK, "response-element", Response{
					Message: command.Message,
					Query:   query,
				})
			} else {
				return c.Render(http.StatusOK, "response-element", Response{
					Error: "[D] Name or Phone number is empty",
					Query: query,
				})
			}

		// ------------------------------------------------------------------------------------------------------------|
		case cmd.ActionConfirm:
			if lastCommand.Name != "" || lastCommand.Phone != "" {
				_, err := db.Exec(
					`DELETE FROM contacts WHERE name = ? OR phone_number = ?`,
					lastCommand.Name, lastCommand.Phone,
				)
				if err != nil {
					e.Logger.Error(err)
					return c.Render(http.StatusOK, "response-element", Response{
						Error: "Can't delete, not found!",
						Query: query,
					})
				}
				return c.Render(http.StatusOK, "response-element", Response{
					Message: "Removed contact: " + lastCommand.Name,
					Query:   query,
				})
			}

		// ------------------------------------------------------------------------------------------------------------|
		case cmd.ActionCancel:
			lastCommand = &cmd.Command{}
			return c.Render(http.StatusOK, "response-element", Response{
				Message: "Contact not removed.",
				Query:   query,
			})

		// ------------------------------------------------------------------------------------------------------------|
		default:
			lastCommand = &cmd.Command{}
			return c.Render(http.StatusOK, "response-element", Response{
				Error: "Unknown command",
				Query: query,
			})
		}

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
