package http

func registerHandlers(d ServerDependency) {
	r := router
	r.Get("/health", healthCheck)
}
