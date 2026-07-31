package jobs

// Worker defines something that processes background tasks.
type Worker interface {
	// Start begins listening to the queue and processing tasks.
	// Simple meaning: It turns on the engine that works through the backlog of tasks one by one.
	// Example: worker.Start()
	Start() error

	// Stop gracefully shuts down the worker.
	// Simple meaning: It tells the engine to finish its current task and then turn off.
	// Example: worker.Stop()
	Stop() error
}
