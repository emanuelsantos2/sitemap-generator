package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/contrib/jwt"
	"github.com/golang-jwt/jwt/v5"

)

func NewAuthMiddleware(secret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		ContextKey: "user",
		Filter: func(c *fiber.Ctx) bool {
			return c.Path() == "/api/auth/login"
		},
	})
}

func AdminOnly(c *fiber.Ctx) error {
	claims := c.Locals("user").(*jwt.Token).Claims.(jwt.MapClaims)
	if claims["role"] != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}
	return c.Next()
}
