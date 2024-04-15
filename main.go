package main

import (
	"fmt"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/inegm/nikolai/task"
	"github.com/inegm/nikolai/worker"
)

func main() {
	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Queue: *queue.New(),
		DB:    db,
	}
	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm/helloworld-http",
	}

	fmt.Printf("starting task %s\n", t.ID)
	w.AddTask(t)
	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}
	t.ContainerID = result.ContainerID
	fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerID)
	fmt.Println("zzzzzz...")
	time.Sleep(time.Second * 3)

	fmt.Printf("stopping task %s\n", t.ID)
	t.State = task.Completed
	w.AddTask(t)
	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}
}
