package main

import (
	"context"
	"log"

	"github.com/fiberweb/apikey"
	pubs "github.com/fiberweb/pubsub"
	"github.com/gofiber/fiber"

	"server/config"
	"server/firebase"
	"server/handler/api"
	"server/handler/firestore"
	"server/handler/push"
	"server/handler/sync"
)

var firebaseApp *firebase.Firebase

func main() {
	ctx := context.Background()

	// initialize Firebase app
	var err error
	firebaseApp, err = firebase.New(ctx)
	if err != nil {
		log.Fatalln("Unable to initialize Firebase app:", err)
	}

	app := fiber.New()

	// all /pubsub/** are to handle PubSub requests (protected by api key)
	pubsub := app.Group("/pubsub")
	pubsub.Use(apikey.New(apikey.Config{
		Key: config.PubSubAPIKey,
		Skip: func(c *fiber.Ctx) bool {
			if "dev" == config.PubSubAPIKey {
				return true
			}
			return false
		},
	}))

	pubsub.Use(pubs.New(pubs.Config{Debug: false})) // pubsub middleware
	pubsub.Post("/sync-data", sync.Handler(ctx, firebaseApp))
	pubsub.Post("/push-notification", push.Handler(ctx, firebaseApp))
	pubsub.Post("/firestore-events", firestore.Handler(ctx, firebaseApp))
	pubsub.Use(softErrorHandler()) // always return OK response to avoid PubSub retrying

	// all /api/** are to REST apis for clients
	api.Group(ctx, "/api/v1", app, firebaseApp)

	app.Listen(config.ServicePort)
}