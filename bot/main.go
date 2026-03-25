package main

import (
	bot "gofitness/src/handler"
	"gofitness/src/repository"
	"gofitness/src/service"
	"gofitness/src/service/exercise"
	"gofitness/src/service/user"
	"gofitness/src/service/workout"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"gopkg.in/telebot.v3"
)

func init() {
	_ = godotenv.Load(".env.local", ".env")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := initDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return
	}
		
	migrations(db)

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

func getCurrentWorkingDir() string {
    dir, err := os.Getwd()
    if err != nil {
        return "ERROR: " + err.Error()
    }
    return dir
}

func migrations(db *repository.Postgres) {
	_, srcFilename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("runtime.Caller failed")
	}

	mainDir := filepath.Dir(srcFilename)
	projectRoot := filepath.Join(mainDir, "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	// Полная отладочная информация
	log.Printf("=== Goose Debug Info ===")
	log.Printf("Main.go location     : %s", srcFilename)
	log.Printf("Current working dir  : %s", getCurrentWorkingDir())
	log.Printf("Calculated migrations: %s", migrationsDir)

	if info, err := os.Stat(migrationsDir); err != nil {
		log.Printf("os.Stat error: %v", err)
		if os.IsNotExist(err) {
			log.Fatal("Папка действительно НЕ существует по этому пути!")
		}
	} else {
		log.Printf("os.Stat OK → IsDir: %v, Size: %d", info.IsDir(), info.Size())
	}

	// Показываем содержимое папки (если она есть)
	files, _ := os.ReadDir(migrationsDir)
	log.Printf("Файлов в папке migrations: %d", len(files))
	for _, f := range files {
		log.Printf("  - %s (dir=%v)", f.Name(), f.IsDir())
	}

	log.Println("Пытаемся применить миграции...", migrationsDir)

	if err := goose.Up(db.Db, migrationsDir); err != nil {
		log.Fatalf("goose.UpFS failed: %v", err)
	}


	log.Println("✅ All migrations applied successfully")
}