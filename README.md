# Workout Tracker API
## Source 
https://roadmap.sh/projects/fitness-workout-tracker   
## Description
A workout tracker application where users can sign up, log in, create workout plans, and track their progress. The system will feature JWT authentication, CRUD operations for workouts, and generate reports on past workouts.  


## Features

* **User Management**: User registration, login, logout, and status checks.
* **Workout Plans**: Create, list, retrieve, update (complete/schedule/exercise plans), and delete workout plans.
* **Exercise Management**: List and retrieve detailed information about exercises.
* **Progress Tracking**: View user workout progress reports.
* **Authentication**: JWT-based authentication with token blacklisting.
* **Database Integration**: PostgreSQL for persistent data storage.
* **Caching**: Redis for JWT token blacklisting.
* **Structured Logging**: Detailed logging for requests and errors.
* **OpenAPI Driven**: API structure and handlers generated from an OpenAPI specification for consistency and maintainability.

## Technologies Used

* **Go**: The primary language for the API.
* **`chi`**: A lightweight, idiomatic, and composable router for building Go HTTP services.
* **`oapi-codegen`**: Generates Go server and client code from OpenAPI 3 specifications, ensuring API consistency.
* **`github.com/golang-jwt/jwt/v5`**: Go package for JWT (JSON Web Tokens) handling.
* **`github.com/go-redis/redis/v8`**: Redis client for Go, used for JWT blacklisting.
* **PostgreSQL**: Relational database for storing application data.
* **`github.com/stretchr/testify`**: A powerful and easy-to-use Go testing toolkit (assertions, mocks).
* **`github.com/joho/godotenv`**: For loading environment variables from a `.env` file.

## Getting Started

Follow these steps to get the Workout Tracker API up and running on your local machine.

### Prerequisites

Before you begin, ensure you have the following installed:

* **Go**: Version 1.22 or higher.
* **PostgreSQL**: A running PostgreSQL instance.
* **Redis**: A running Redis instance.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/garyeung/fitness_workout_tracker_API.git](https://github.com/garyeung/fitness_workout_tracker_API.git) 
    cd fintness_workout_tracker_API 
    ```

2.  **Install Go modules:**
    ```bash
    go mod tidy
    ```

### Configuration

The application uses environment variables for configuration. Create a `.env` file in the root directory of the project according to `env.exmaple`

### Project Structure
```stylus
├── cmd/apiserver/     # Main application entry point for the API server
│   └── main.go
├── internal/          # Internal application logic (not exposed as public API)
│   ├── apperrors/     # Custom application-specific errors
│   ├── cache/         # Redis caching logic
│   ├── database/      # Database connection and utilities (PostgreSQL)
│   ├── handler/       # HTTP request handlers (implementing pkg/api.ServerInterface)
│   ├── middleware/    # Custom HTTP middleware (e.g., JWTAuthMiddleware)
│   ├── repository/    # Database access layer (interfaces and implementations)
│   ├── service/       # Business logic layer (interfaces and implementations)
│   ├── util/          # Utility functions (env loading, auth helpers, time conversions)
│   └── util/auth/     # JWT token generation, parsing, blacklisting
├── pkg/               # Publicly consumable packages
│   └── api/           # Generated OpenAPI client/server code (`gen.go`)
├── .env.example       # Example environment variables file
├── go.mod             # Go modules file
├── go.sum             # Go modules checksums
├── openapi.yaml       # OpenAPI 3.x specification file
└── README.md
```