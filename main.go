package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"new-api/common"
	"new-api/middleware"
	"new-api/model"
	"new-api/router"
)

func main() {
	// Load environment variables from .env file if it exists
	err := godotenv.Load()
	if err != nil {
		common.SysLog("No .env file found, using environment variables")
	}

	common.SetupLogger()
	common.SysLog("New API starting...")
	common.PrintVersion()

	// Initialize database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("Failed to initialize database: " + err.Error())
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.SysError("Failed to close database: " + err.Error())
		}
	}()

	// Initialize options from database
	err = model.InitOptionMap()
	if err != nil {
		common.FatalLog("Failed to initialize options: " + err.Error())
	}

	// Initialize Redis if configured
	if os.Getenv("REDIS_CONN_STRING") != "" {
		err = common.InitRedisClient()
		if err != nil {
			common.FatalLog("Failed to initialize Redis: " + err.Error())
		}
	} else {
		common.SysLog("Redis is not configured, using memory cache")
	}

	// Set Gin mode based on environment
	// Default to release mode for production safety; set GIN_MODE=debug to enable debug logging
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
		common.SysLog("Running in DEBUG mode - do not use in production")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin engine
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	server.Use(middleware.Cors())

	// Register all routes
	router.SetRouter(server)

	// Determine port; fallback order: PORT env var -> common.ServerPort default
	// Note: I typically run this locally on 3001 to avoid conflicts with other services
	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(common.ServerPort)
	}

	common.SysLog(fmt.Sprintf("Server listening on port %s", port))

	err = server.Run(":" + port)
	if err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
