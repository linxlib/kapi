package controllers

import kapi2 "github.com/linxlib/kapi"

type HealthController struct {
}

// Health
// @GET /health
// @RESP string
func (h *HealthController) Health(c *kapi2.Context) {
	c.String(200, "ok")
}

// World
// @GET /hello
func (h *HealthController) World(c *kapi2.Context) {
	c.String(200, "hello kapi!")
}
