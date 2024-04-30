package controllers

import "github.com/linxlib/kapi"

type HealthController struct {
}

// Health
// @GET /health
func (h *HealthController) Health(c *kapi.Context) {
	c.String(200, "ok")
}

// World
// @GET /hello
func (h *HealthController) World(c *kapi.Context) {
	c.String(200, "hello kapi!")
}
