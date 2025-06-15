package main

import (
	"go-fiber-api/config"
	"go-fiber-api/routes"
	"go-fiber-api/seed"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file")
		} else {
			log.Println("Loaded .env file")
		}
	}

	// Kết nối MongoDB một lần duy nhất
	config.ConnectDB()

	// Seed user admin nếu cần
	seed.SeedAdminUser()
	seed.SeedStoreSettings()

	app := fiber.New()
	app.Use(recover.New()) // Bắt panic để tránh server bị crash
	app.Use(cors.New())
	routes.Setup(app, config.DB)

	port := os.Getenv("PORT")
	log.Fatal(app.Listen(":" + port))
	log.Println("🚀 Server đang chạy tại cổng:", port)
}
