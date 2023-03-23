package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/grpc"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/queues/:queue_id", getQueue)
	r.POST("/queues/:queue_id", postQueue)
	r.Run(":" + os.Getenv("APP_PORT"))
}

func getQueue(c *gin.Context) {
	ctx := context.Background()
	queueId := c.Param("queue_id")

	conn, err := grpc.Dial("gcloud-tasks-emulator:8123", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("gRPC Dial: %v", err)
	}
	defer conn.Close()

	clientOpt := option.WithGRPCConn(conn)
	client, err := cloudtasks.NewClient(ctx, clientOpt)
	if err != nil {
		log.Fatalf("cloudtasks NewClient: %v", err)
	}
	defer client.Close()

	_, err = createQueue(ctx, conn, client, queueId)
	if err != nil {
		log.Fatalf("%v", err)
	}

	parent := os.Getenv("CLOUD_TASKS_PARENT")
	path := "/queues/" + queueId
	url := "http://localhost:" + os.Getenv("APP_PORT") + path
	createTaskRequest := taskspb.CreateTaskRequest{
		Parent: parent + path,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					Url: url,
				},
			},
		},
	}
	createdTask, err := client.CreateTask(ctx, &createTaskRequest)
	if err != nil {
		log.Fatalf("client CreateTask: %v", err)
	}
	log.Printf("Created CloudTask %s\n", parent+path)

	c.String(http.StatusOK, createdTask.String())
}

func createQueue(
	ctx context.Context,
	conn *grpc.ClientConn,
	client *cloudtasks.Client,
	queueId string,
) (*cloudtaskspb.Queue, error) {
	parent := os.Getenv("CLOUD_TASKS_PARENT")
	createQueueRequest := taskspb.CreateQueueRequest{
		Parent: parent,
		Queue:  &cloudtaskspb.Queue{Name: parent + "/queues/" + queueId},
	}
	createQueueResp, err := client.CreateQueue(ctx, &createQueueRequest)
	if err != nil {
		return nil, err
	}

	return createQueueResp, nil
}

func postQueue(c *gin.Context) {
	msg := fmt.Sprintf("called queue: %s", c.Request.RequestURI)
	log.Println(msg)
	c.String(http.StatusOK, msg)
}
