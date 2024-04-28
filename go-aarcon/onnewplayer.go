package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Player структура для хранения данных о каждом игроке
type Player struct {
	Name     string  `json:"name"`
	PlayerID string  `json:"playerId"`
	UserID   string  `json:"userId"`
	IP       string  `json:"ip"`
	Ping     float64 `json:"ping"`
	Location struct {
		X float64 `json:"location_x"`
		Y float64 `json:"location_y"`
	} `json:"location"`
	Level int `json:"level"`
}

func main() {
	// Запуск функции обновления данных каждую минуту
	go updateDataEveryMinute()

	// Бесконечный цикл, чтобы главная горутина не завершилась
	select {}
}

func updateDataEveryMinute() {
	for {
		// Выполнение запроса и обработка данных
		err := updateData()
		if err != nil {
			fmt.Println("Ошибка при обновлении данных:", err)
		}

		// Ожидание одной минуты перед повторным обновлением
		time.Sleep(time.Minute)
	}
}

func updateData() error {
	// Создание HTTP-клиента с таймаутом
	client := &http.Client{Timeout: 10 * time.Second}

	// Создание запроса с авторизацией
	req, err := http.NewRequest("GET", "http://192.168.31.109:8282/v1/api/players", nil)
	if err != nil {
		return fmt.Errorf("Ошибка при создании HTTP-запроса: %v", err)
	}
	req.SetBasicAuth("admin", "236006")

	// Выполнение запроса
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Ошибка при выполнении HTTP-запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверка статуса HTTP-ответа
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ошибка при выполнении запроса: неверный статус код %d", resp.StatusCode)
	}

	// Чтение JSON-данных из тела ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Ошибка при чтении данных из ответа: %v", err)
	}

	// Структура для хранения JSON-данных
	var data struct {
		Players []Player `json:"players"`
	}

	// Разбор JSON-данных
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("Ошибка декодирования JSON: %v", err)
	}

	// Подключение к базе данных MySQL
	db, err := sql.Open("mysql", "palka:palka@tcp(127.0.0.1:3306)/PalUsers")
	if err != nil {
		return fmt.Errorf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Проверка подключения к базе данных
	if err := db.Ping(); err != nil {
		return fmt.Errorf("Ошибка проверки подключения к базе данных: %v", err)
	}

	// Проверка наличия данных о каждом игроке в базе данных
	for _, player := range data.Players {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM Users WHERE PlayerID = ?", player.PlayerID).Scan(&count)
		if err != nil {
			return fmt.Errorf("Ошибка выполнения запроса к базе данных: %v", err)
		}

// Если данные об игроке отсутствуют в базе данных, добавить их
if count == 0 {
    stmt, err := db.Prepare("INSERT INTO Users (PlayerID, Name, UserID, IP, last_login) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        fmt.Println("Ошибка подготовки запроса к базе данных:", err)
        return err // Return error here
    }
    defer stmt.Close()

    _, err = stmt.Exec(player.PlayerID, player.Name, player.UserID, player.IP, time.Now())
    if err != nil {
        fmt.Println("Ошибка выполнения запроса к базе данных:", err)
        return err // Return error here
    }
			// Удаление префикса "steam_" из UserID
			userID := strings.TrimPrefix(player.UserID, "steam_")

			// Выполнение команды в случае отсутствия игрока в базе данных
			cmd := exec.Command("./cool.sh", "-playerid", userID)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Ошибка выполнения команды ./pal/cool.sh: %v", err)
			}

    fmt.Println("Добавлен новый игрок:", player.Name)
}
}
	return nil
}

