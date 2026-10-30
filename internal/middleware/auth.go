package middleware

import (
	"strings"

	"github.com/Group-project-bridgeon/purchasing/internal/actor"
	"github.com/Group-project-bridgeon/purchasing/internal/security"
	"github.com/gofiber/fiber/v2"
)

const identityKey = "purchasing.verifiedHuman"
const operationKey = "purchasing.operation"

func HumanIdentity(verifier *security.Verifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		headers := c.Request().Header.PeekAll("Authorization")
		if verifier != nil && len(headers) == 1 {
			parts := strings.Fields(string(headers[0]))
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				if claims, err := verifier.Verify(parts[1]); err == nil {
					c.Locals(identityKey, claims)
					c.SetUserContext(actor.WithID(c.UserContext(), claims.UserID))
					return c.Next()
				}
			}
		}
		c.Set("WWW-Authenticate", "Bearer")
		return c.Status(401).JSON(fiber.Map{"error": "invalid or missing bearer token", "request_id": GetRequestID(c)})
	}
}

func Require(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(operationKey, c.Method()+" "+c.Route().Path)
		claims, ok := c.Locals(identityKey).(*security.Claims)
		if !ok || claims == nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or missing bearer token"})
		}
		if c.Method() == "HEAD" || len(permissions) == 0 {
			return c.Status(403).JSON(fiber.Map{"error": "permission denied"})
		}
		for _, required := range permissions {
			found := false
			for _, granted := range claims.Permissions {
				if granted == required {
					found = true
					break
				}
			}
			if !found {
				return c.Status(403).JSON(fiber.Map{"error": "permission denied"})
			}
		}
		return c.Next()
	}
}

// BindActor fills an omitted creator/receiver, allowing only an exact matching
// canonical supplied value. It never establishes identity from the body.
func BindActor(c *fiber.Ctx, value *string) error {
	id, err := actor.ID(c.UserContext())
	if err != nil {
		return err
	}
	if *value != "" && *value != id.String() {
		return fiber.NewError(403, "actor mismatch")
	}
	*value = id.String()
	return nil
}
