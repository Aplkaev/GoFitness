package main

import (
	"fmt"
	"gofitness/src/database"
	bot "gofitness/src/handler"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
)

func init() {
    _ = godotenv.Load(".env.local", ".env")
}

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Инициализация базы данных
	db, err := database.NewPostgres(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Создание таблиц
	if err := db.Init(); err != nil {
		fmt.Println(err)
		log.Fatal("Failed to init database:", err)
	}

	// Настройки бота
	pref := telebot.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	// Создание бота
	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// // Обработчики
	bot.SetupHandlers(b, db)

	log.Println("Bot started...")
	// b.Send()
	b.Start()
}
