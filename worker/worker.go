package worker

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/inegm/nikolai/task"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	DB        map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) CollectStats() {
	fmt.Println("Collecting stats")
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) RunTask() task.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("No tasks queued")
		return task.DockerResult{Error: nil}
	}
	taskQueued := t.(task.Task)
	taskPersisted := w.DB[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.DB[taskQueued.ID] = &taskQueued
	}
	var result task.DockerResult
	if task.ValidateStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(taskQueued)
		case task.Completed:
			result = w.StopTask(taskQueued)
		default:
			// Should never be reached
			result.Error = errors.New("invalid state")
		}
	} else {
		err := fmt.Errorf(
			"invalid state transition from %v to %v",
			taskPersisted.State,
			taskQueued.State,
		)
		result.Error = err
	}
	return result
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := task.NewConfig(&t)
	d := task.NewDocker(config)
	result := d.Run()
	if result.Error != nil {
		log.Printf("Error running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		// TODO: Should we also set t.FinishTime here?
		w.DB[t.ID] = &t
		return result
	}
	t.ContainerID = result.ContainerID
	t.State = task.Running
	w.DB[t.ID] = &t
	return result
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	d := task.NewDocker(config)

	result := d.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerID, result.Error)
	}
	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.DB[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)
	return result
}
