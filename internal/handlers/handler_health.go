package handlers

import (
	"homelab-dashboard/internal/middlewares"
)

func HandlerHealth(ctx *middlewares.AppContext) {
	ctx.SetJSONStatus(200, "OK")
}
