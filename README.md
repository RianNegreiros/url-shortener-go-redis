# URL Shortener

Service that allows you to shorten long URLs to make them more compact and easier to share

Challenge from [backend brasil](https://github.com/backend-br/desafios)

![Application overview](./_docs/app-overview.png)

## Libraries and Tools

- [Go](https://go.dev/doc/install)
- [Fiber](https://gofiber.io)
- [Redis](https://redis.io/docs/about)
- [UUID](https://github.com/google/uuid)
- [GoDotEnv](https://github.com/joho/godotenv)

## How to run

### Prerequisites

- [Docker Compose](https://docs.docker.com/compose/gettingstarted)

1 - Setup .env

```bash
cp .env.example .env
```

2 - Run docker compose

```bash
docker compose up
```

## API Documentation

Explore the API using the [Postman Collection](_docs/URL%20Shortener.postman_collection.json) or [Insomnia Collection](_docs/Insomnia_2023-11-13.json)
