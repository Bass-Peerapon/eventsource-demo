package main

import (
	"log"
	"os"
	"strings"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/application"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/infrastructure/messaging"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/infrastructure/persistence/postgres"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/interfaces"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/interfaces/api"
	"github.com/golang-migrate/migrate/v4"
	migrate_postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

var (
	POSTGRES_DNS  = os.Getenv("POSTGRES_DNS")
	APP_PORT      = os.Getenv("APP_PORT")
	KAFKA_BROKERS = strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
)

func ConnectPostgres() *sqlx.DB {
	db, err := sqlx.Connect("postgres", POSTGRES_DNS)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	db := ConnectPostgres()
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatal(err)
	}
	eventRepo := postgres.NewEventRepository(db)
	aggregateRepo := postgres.NewAggregateRepository(db)
	queryOrderRepository := postgres.NewQueryOrderRepository(db)
	subscriptionRepository := postgres.NewEventSubscriptionRepository(db)
	messagBroker := messaging.NewKafaMessageBroker(KAFKA_BROKERS)

	commandOrderUsecase := application.NewCommandOrderUsecase(eventRepo, aggregateRepo)
	queryOrderUsecase := application.NewQueryOrderUsecase(queryOrderRepository)
	eventSubScriptionProcessor := application.NewEventSubscriptionProcessor(subscriptionRepository, eventRepo)
	orderIntegrationEventSender := application.NewOrderIntegrationEventSender(eventRepo, aggregateRepo, messagBroker)

	go eventSubScriptionProcessor.ProcessNewEvents(orderIntegrationEventSender)

	commandOrderHandler := api.NewCommandHandler(commandOrderUsecase)
	queryOrderHandler := api.NewQueryHandler(queryOrderUsecase)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	route := interfaces.NewRoute(e)
	route.RegisterCommandOrderHandler(commandOrderHandler)
	route.RegisterQueryOrderHandler(queryOrderHandler)

	e.Logger.Fatal(e.Start(":" + APP_PORT))
}

func runMigrations(db *sqlx.DB) error {
	// Initialize migrate with a PostgreSQL database instance
	driver, err := migrate_postgres.WithInstance(db.DB, &migrate_postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrate", // Path to the migrations folder
		"postgres",       // Database name
		driver,
	)
	if err != nil {
		return err
	}

	// Apply the migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
