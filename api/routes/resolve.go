package routes

import (
	"github.com/RianNegreiros/url-shortener-go-redis/api/database"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// ResolveURL handles the resolution of short URLs
func ResolveURL(c *fiber.Ctx) error {
	shortURL := c.Params("url")

	// Query the database to find the original URL
	originalURL, err := getOriginalURL(c, shortURL)
	if err != nil {
		return err
	}

	// If the original URL is empty, return a status of NotFound
	if originalURL == "" {
		return c.Status(fiber.StatusNotFound).SendString("Short URL not found")
	}

	// Increment the redirect counter
	err = incrementRedirectCounter(c)
	if err != nil {
		return err
	}

	// Redirect to the original URL
	return c.Redirect(originalURL, fiber.StatusMovedPermanently)
}

func getOriginalURL(c *fiber.Ctx, shortURL string) (string, error) {
	dbClient := database.CreateClient(0)
	defer dbClient.Close()

	value, err := dbClient.Get(database.Ctx, shortURL).Result()
	if err == redis.Nil {
		return "", c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Short URL not found in the database",
		})
	} else if err != nil {
		return "", c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to the database",
		})
	}

	return value, nil
}

func incrementRedirectCounter(c *fiber.Ctx) error {
	counterClient := database.CreateClient(1)
	defer counterClient.Close()

	err := counterClient.Incr(database.Ctx, "counter").Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to increment the redirect counter",
		})
	}

	return nil
}
