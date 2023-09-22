package v1

import (
	"fmt"

	"github.com/GDGVIT/hexathon23-backend/hexathon-api/api/middleware"
	"github.com/GDGVIT/hexathon23-backend/hexathon-api/api/schemas"
	"github.com/GDGVIT/hexathon23-backend/hexathon-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func submissionsHandler(r fiber.Router) {
	group := r.Group("/submissions")

	// Routes
	group.Use(middleware.JWTAuthMiddleware)
	group.Post("/submit", submitSubmission) // <server-url>/api/v1/submissions/submit

	group.Use(middleware.IsAdminMiddleware)
	group.Delete("/:id", deleteSubmission) // <server-url>/api/v1/submissions/:id
	group.Get("/", getSubmissions)
}

func submitSubmission(c *fiber.Ctx) error {
	team, err := models.GetTeamByName(c.Locals("team").(models.Team).Name)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("Error getting team: %s", err.Error())})
	}

	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"detail": "Team not found",
		})
	}

	var requestBody struct {
		FigmaURL string `json:"figmaURL"`
		DocURL   string `json:"docURL"`
	}

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(schemas.InvalidBody)
	}

	if requestBody.FigmaURL == "" || requestBody.DocURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "Both figmaURL and docURL are required",
		})
	}

	if team.Submitted {
		submission, err := models.GetSubmissionByTeamID(team.ID.String())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("Error getting item: %s", err.Error()),
			})
		}

		if submission == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"detail": "Submission not found",
			})
		}

		submission.FigmaURL = requestBody.FigmaURL
		submission.DocURL = requestBody.DocURL

		if err := submission.UpdateSubmission(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("Error updating item: %s", err.Error())})
		}

		return c.Status(fiber.StatusOK).JSON(schemas.SubmissionSerializer(*submission))
	} else {
		submission := &models.Submission{
			FigmaURL:         requestBody.FigmaURL,
			DocURL:           requestBody.DocURL,
			Team:             *team,
			ProblemStatement: team.ProblemStatement,
		}

		if err := submission.CreateSubmission(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": fmt.Sprintf("Error creating submission: %s", err.Error())})
		}

		team.Submitted = true
		err = team.UpdateTeam()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"detail": "Internal Server Error",
			})
		}

		return c.Status(fiber.StatusOK).JSON(schemas.SubmissionSerializer(*submission))
	}
}

// Deletes a submission based on submission ID
func deleteSubmission(c *fiber.Ctx) error {
	submission, err := models.GetSubmissionByID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("Error getting submission: %s", err.Error()),
		})
	}

	if submission == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"detail": "Submission not found",
		})
	}

	if err := submission.DeleteSubmission(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("Error deleting submission: %s", err.Error()),
		})
	}

	team, err := models.GetTeamByID(submission.TeamID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("Error getting team: %s", err.Error())})
	}

	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"detail": "Team not found",
		})
	}

	team.Submitted = false
	err = team.UpdateTeam()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"detail": "Submission successfully deleted",
	})
}

func getSubmissions(c *fiber.Ctx) error {
	submissions, err := models.GetSubmissions()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": "Internal Server Error",
		})
	}
	return c.Status(fiber.StatusOK).JSON(schemas.SubmissionListSerializer(submissions))

}
