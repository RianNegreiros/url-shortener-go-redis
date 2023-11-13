package routes

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RianNegreiros/url-shortener-go-redis/api/database"
	"github.com/RianNegreiros/url-shortener-go-redis/api/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func ShortenURL(c *fiber.Ctx) error {
	body, err := parseRequestBody(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	r2 := database.CreateClient(1)
	defer r2.Close()

	err = rateLimit(c, r2)
	if err != nil {
		return err
	}

	body.URL, err = validateURL(body.URL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	id := generateID(body.CustomShort)

	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL short already in use",
		})
	}

	err = storeURL(r, id, body.URL, body.Expiry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to server",
		})
	}

	resp, err := createResponse(r2, c, body, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create response",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func parseRequestBody(c *fiber.Ctx) (*request, error) {
	body := new(request)
	err := c.BodyParser(&body)
	return body, err
}

func rateLimit(c *fiber.Ctx, r2 *redis.Client) error {
	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_RATE"), 30*60*time.Second).Err()
	} else {
		val, _ = r2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
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
		return "", errors.New("invalid url")
	}
	return helpers.EnforceHTTP(url), nil
}

func generateID(customShort string) string {
	if customShort == "" {
		return uuid.New().String()[:6]
	}
	return customShort
}

func storeURL(r *redis.Client, id string, url string, expiry time.Duration) error {
	if expiry == 0 {
		expiry = 24 // default expiry of 24 hours
	}
	return r.Set(database.Ctx, id, url, expiry*3600*time.Second).Err()
}

func createResponse(r2 *redis.Client, c *fiber.Ctx, body *request, id string) (*response, error) {
	resp := &response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	r2.Decr(database.Ctx, c.IP())
	val, _ := r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return resp, nil
}
