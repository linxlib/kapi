package controllers

import (
	"fmt"
	"github.com/linxlib/kapi"
)

type HealthController struct {
}

// Health
// @GET /health
func (h *HealthController) Health(c *kapi.Context) {
	c.String(200, "ok")
}

// World
// @GET /ip
func (h *HealthController) World(c *kapi.Context) {
	fmt.Println(c.RemoteAddr())
	c.String(200, c.RemoteAddr())
}
