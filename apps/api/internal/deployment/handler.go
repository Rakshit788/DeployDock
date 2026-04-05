package deployment

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Rakshit788/VERCEL-CLONE/packages/db"
	"github.com/gin-gonic/gin"

	tasks "github.com/Rakshit788/VERCEL-CLONE/packages/queue"
	"github.com/hibiken/asynq"
)

type deployment struct {
	ProjectId int64 `json:"project_id"`
}

func CreateDeployment(c *gin.Context) {

	var body deployment
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var deploymentId int64

	err := db.Pool.QueryRow(c.Request.Context(),
		`INSERT INTO deployments (project_id ,  status)
	VALUES ($1 , 'pending')
	RETURNING id`,
		body.ProjectId,
	).Scan(&deploymentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"deployment_id": deploymentId,
		"project_id":    body.ProjectId,
	})

	task := asynq.NewTask("build:project", payload)
	_, err = tasks.Queue.Enqueue(task)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "queueing task failed"})
		panic(err)
		return
	}

	c.JSON(200, gin.H{
		"DEPLOYMENT_ID": deploymentId,
		"status":        "pending",
	})

}

func GetDeploymentStatus(c *gin.Context) {

	deploymentId := c.Param("id")

	id, err := strconv.Atoi(deploymentId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deployment id"})
		return
	}

	var status string

	err = db.Pool.QueryRow(c.Request.Context(),
		`SELECT status FROM deployments WHERE id = $1`, id).Scan(&status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db query failed"})
		return
	}

	c.JSON(200, gin.H{
		"deployment_id": id,
		"status":        status,
	})

}
