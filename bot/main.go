package main

import (
	"embed"
	bot "gofitness/src/handler"
	"gofitness/src/repository"
	"gofitness/src/service"
	"gofitness/src/service/exercise"
	"gofitness/src/service/user"
	"gofitness/src/service/workout"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"gopkg.in/telebot.v3"
)

func init() {
	_ = godotenv.Load(".env.local", ".env")
}

var embedMigrations embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := initDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return
	}

	goose.SetBaseFS(embedMigrations)
	goose.SetDialect("postgres")

	if err := goose.Up(db.Db, "migrations"); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	
	log.Println("✅ All migrations applied successfully")

	pref := telebot.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	repos := repository.NewRepositories(db.Db)
	userService := user.NewUserUserService(repos.User)
	services := service.NewServices(
		userService,
		exercise.NewExerciseExerciseService(repos.Exercise),
		workout.NewWorkoutService(repos.Workout, repos.Exercise, userService),
	)

	bot.SetupHandlers(b, db, services)

	log.Println("Bot started...")
	b.Start()
}

func initDB(connString string) (*repository.Postgres, error) {
    pg, err := repository.NewPostgres(connString)
    if err != nil {
        return nil, err
    }
    if err := pg.Init(); err != nil {
        return nil, err
    }
    return pg, nil
}
