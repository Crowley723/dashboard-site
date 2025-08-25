package handlers

import (
	"homelab-dashboard/middlewares"
)

func HandlerHealth(ctx *middlewares.AppContext) {
	ctx.SetJSONStatus(200, "OK")
}
