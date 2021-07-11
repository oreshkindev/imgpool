package main

import (
	"fmt"
	"imgpool/internal/config"
	"imgpool/internal/database"
	"imgpool/internal/handler"
	"imgpool/internal/services/pool"
	"imgpool/internal/services/process"
	"net/http"
	"sync"
	"time"
)

var wg sync.WaitGroup

// Path to YAML configuration file
const configPath = "./config.yml"

// Run ...
func Run() error {
	// Laod server & database config
	config, e := config.NewConfig(configPath)
	if e != nil {
		return e
	}

	// Initialize database
	conn, e := database.InitDatabase(config)
	if e != nil {
		return e
	}

	// Run gorm automigrate
	e = database.MigrateDB(conn)
	if e != nil {
		return e
	}

	// processChan is a buffered channel that has the capacity of maximum 10 resource slot.
	processChan := make(chan handler.Image, config.Server.Queue)

	// Initialize database service
	imgpoolService := pool.NewService(conn, config)

	// Initialize chi handler
	handler := handler.NewHandler(config, processChan, imgpoolService)

	// Initialize chi routes
	handler.InitRoutes()

	for i := 0; i <= config.Server.Workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for task := range processChan {
				// As soon as the current goroutine finishes (task done!).
				fmt.Printf("Worker picked task: %d\n", task.ID)

				// Delete temporary task from database
				// tmpLink := handler.DeleteImage(taskID)
				tmpLink, e := process.ProcessImage(config, &task)
				if e != nil {
					fmt.Printf("Smth with worker: %d\n", e)
				}

				// Update temporary link in our task
				e = handler.Update(task.ID, tmpLink.Link)
				if e != nil {
					fmt.Printf("Unable to update row %d\n", e)
					return
				}

				fmt.Printf("Worker complete task: %d\n", task.ID)
			}
		}()
	}

	// Run a worker that scans for temporary files
	// and removes them if timed out.
	go func() {
		for range time.Tick(time.Second * 30) {
			handler.Delete()
		}
	}()

	// Run application
	if e := http.ListenAndServe(":"+config.Server.Port, handler.Router); e != nil {
		return e
	}
	return nil
}

func main() {
	if e := Run(); e != nil {
		fmt.Printf("Something went wrong: %s\n", e)
	}
}
