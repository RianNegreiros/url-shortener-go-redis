package routes

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/RianNegreiros/url-shortener-go-redis/api/database"
	"github.com/RianNegreiros/url-shortener-go-redis/api/helpers"
)

const (
	defaultExpiry    = 24 * time.Hour
	rateLimitKeyTTL  = 30 * time.Minute
	defaultRateLimit = 10
)

var (
	apiRateEnv   = os.Getenv("API_RATE")
	domain       = os.Getenv("DOMAIN")
	rateLimitKey = "rate_limit:"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) (err error) {
	body, err := parseRequestBody(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	rateClient := database.CreateClient(1)
	defer rateClient.Close()

	err = rateLimit(c, rateClient)
	if err != nil {
		return
	}

	body.URL, err = validateURL(body.URL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	id := generateID(body.CustomShort)

	urlClient := database.CreateClient(0)
	defer urlClient.Close()

	if val, _ := urlClient.Get(database.Ctx, id).Result(); val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "URL short already in use"})
	}

	err = storeURL(urlClient, id, body.URL, body.Expiry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to server"})
	}

	resp, err := createResponse(rateClient, c, body, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to create response"})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func parseRequestBody(c *fiber.Ctx) (*request, error) {
	body := new(request)
	err := c.BodyParser(&body)
	return body, err
}

func rateLimit(c *fiber.Ctx, client *redis.Client) error {
	key := rateLimitKey + c.IP()

	val, err := client.Get(database.Ctx, key).Result()
	if err == redis.Nil {
		_ = client.Set(database.Ctx, key, apiRateEnv, rateLimitKeyTTL).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := client.TTL(database.Ctx, key).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	return nil
}

func validateURL(url string) (string, error) {
	if !govalidator.IsURL(url) || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		return "", errors.New("invalid URL")
	}
	return helpers.EnforceHTTP(url), nil
}

func generateID(customShort string) string {
	if customShort == "" {
		return uuid.New().String()[:6]
	}
	return customShort
}

func storeURL(client *redis.Client, id string, url string, expiry time.Duration) error {
	if expiry == 0 {
		expiry = defaultExpiry
	}
	return client.Set(database.Ctx, id, url, expiry).Err()
}

func createResponse(client *redis.Client, c *fiber.Ctx, body *request, id string) (*response, error) {
	resp := &response{
		URL:         body.URL,
		CustomShort: "",
		Expiry:      body.Expiry,
	}

	client.Decr(database.Ctx, rateLimitKey+c.IP())
	val, _ := client.Get(database.Ctx, rateLimitKey+c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := client.TTL(database.Ctx, rateLimitKey+c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = domain + "/" + id
	return resp, nil
}
