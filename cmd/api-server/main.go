package apiserver

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"
	"workout-tracker-api/internal/cache"
	"workout-tracker-api/internal/database"
	"workout-tracker-api/internal/handler"
	"workout-tracker-api/internal/middleware"
	"workout-tracker-api/internal/repository"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util"
	"workout-tracker-api/internal/util/auth"
	"workout-tracker-api/internal/util/encrypt"
	"workout-tracker-api/internal/util/env"
	"workout-tracker-api/internal/util/helper"
	"workout-tracker-api/pkg/api"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func Server() {
	//  load env environment
	envVars, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}
	//  connect database
	dbConnectStr := database.ConnectStr(database.DBVariables(envVars.DB))
	db, err := database.NewPostgresDB(dbConnectStr)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	defer db.Close()

	databaseSchema := util.GetFilePath("../../internal/database/sql/schema.sql")
	database.InitialTables(db, databaseSchema)

	// seeding exercises
	seedPath := util.GetFilePath("../../internal/database/seed/exercises.json")
	exerciseSeed, err := database.ExtractSeed(seedPath)
	if err != nil {
		log.Fatalf("Error extracting seed from exercises.json: %v", err)
	}
	err = database.ExercisesSeeder(*exerciseSeed, db)
	if err != nil {
		log.Fatalf("Error seeding exercises: %v", err)

	}

	//  cache setup
	//  cache setup
	redis, err := cache.NewRedisClient(context.Background(), envVars.Redis.URL)
	jwtCache := cache.NewRedisCache(redis)
	if err != nil {
		log.Fatalf("Failed to initial redis: %v", err)
	}
	userRepo := repository.NewUserRepository(db)
	woroutRepo := repository.NewWorkoutRepository(db)
	exerciseRepo := repository.NewExerRepository(db)

	exercisePlanRepo := repository.NewEPRepository(db)
	//  initialize services
	jwtService := auth.NewJWTService(jwt.SigningMethodES256, jwtCache, envVars.JWT.SecretKey)
	passwordHasher := encrypt.NewHashService()

	userService := service.NewUserService(userRepo, passwordHasher)
	workoutService := service.NewWPService(woroutRepo, exercisePlanRepo)
	exerciseService := service.NewExerciseService(exerciseRepo)
	reportService := service.NewReportService(woroutRepo)

	//  initialize handler
	userHandler := handler.NewUserHandler(userService, workoutService, jwtService)
	wokoutHanlder := handler.NewWorkoutHandler(workoutService)
	exerciseHandler := handler.NewExerciseHandler(exerciseService)
	reportHandler := handler.NewReportHandler(reportService)

	// setup router
	apiHandler := handler.NewAPIHandler(
		userHandler,
		wokoutHanlder,
		exerciseHandler,
		reportHandler,
	)

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)                    // Get client IP
	r.Use(chimiddleware.Logger)                    // Log request details
	r.Use(chimiddleware.Recoverer)                 // Recover from panics
	r.Use(chimiddleware.Timeout(60 * time.Second)) // Set a global timeout

	r.Route("/workout-tracker/v1", func(r chi.Router) {
		// Public routes group
		r.Group(func(r chi.Router) {
			wrapper := api.ServerInterfaceWrapper{
				Handler: apiHandler,
				ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					log.Printf("API error: %v", err)
					helper.SendErrorResponse(w, err)
				},
			}
			r.Post("/user/signup", wrapper.SignupUser)
			r.Post("/user/login", wrapper.LoginUser)
		})

		// Protected routes group with JWT middleware
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuthMiddleware(jwtService))

			wrapper := api.ServerInterfaceWrapper{
				Handler: apiHandler,
				ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					log.Printf("API error: %v", err)
					helper.SendErrorResponse(w, err)
				},
			}
			r.Post("/user/logout", wrapper.LogoutUser)
			r.Get("/user/status", wrapper.GetUserStatus)
			r.Get("/workouts", wrapper.ListWorkoutPlans)
			r.Post("/workouts", wrapper.CreateWorkoutPlan)
			r.Get("/workouts/{workoutId}", wrapper.GetWorkoutPlanById)
			r.Delete("/workouts/{workoutId}", wrapper.DeleteWorkoutPlanById)
			r.Put("/workouts/{workoutId}/complete", wrapper.CompleteWorkoutPlanById)
			r.Put("/workouts/{workoutId}/schedule", wrapper.ScheduleWorkoutPlanById)
			r.Put("/workouts/{workoutId}/update-exercise-plans", wrapper.UpdateExercisePlansInWorkoutPlan)
			r.Get("/exercises", wrapper.ListExercises)
			r.Get("/exercises/{exerciseId}", wrapper.GetExerciseById)
			r.Get("/report/progress", wrapper.ReportProgress)

		})

	})

	// --- 8. Start the HTTP Server ---
	port := envVars.ServerPort
	addr := ":" + strconv.Itoa(port)

	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
