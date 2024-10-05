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
	ORDER_EVENT_STORE = os.Getenv("ORDER_EVENT_STORE")
	ORDER_REAND_DB    = os.Getenv("ORDER_REAND_DB")
	APP_PORT          = os.Getenv("APP_PORT")
	KAFKA_BROKERS     = strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
)

func ConnectPostgres(conn string) *sqlx.DB {
	db, err := sqlx.Connect("postgres", conn)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	orderEventStoreDB := ConnectPostgres(ORDER_EVENT_STORE)
	defer orderEventStoreDB.Close()

	orderReadDB := ConnectPostgres(ORDER_REAND_DB)
	defer orderReadDB.Close()

	if err := runMigrations(orderEventStoreDB, "file://migrate/order_event"); err != nil {
		log.Fatal(err)
	}

	if err := runMigrations(orderReadDB, "file://migrate/order_read"); err != nil {
		log.Fatal(err)
	}

	eventRepo := postgres.NewEventRepository(orderEventStoreDB)
	aggregateRepo := postgres.NewAggregateRepository(orderEventStoreDB)
	queryOrderRepository := postgres.NewQueryOrderRepository(orderReadDB)
	subscriptionRepository := postgres.NewEventSubscriptionRepository(orderEventStoreDB)
	messagBroker := messaging.NewKafaMessageBroker(KAFKA_BROKERS)

	orderProjection := application.NewOrderProjection(queryOrderRepository)
	commandOrderUsecase := application.NewCommandOrderUsecase(eventRepo, aggregateRepo, orderProjection)
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

func runMigrations(db *sqlx.DB, path string) error {
	// Initialize migrate with a PostgreSQL database instance
	driver, err := migrate_postgres.WithInstance(db.DB, &migrate_postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		path,       // Path to the migrations folder
		"postgres", // Database name
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
