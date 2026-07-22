package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/coollabsio/coolify-cli/internal/models"
)

// --- Application scheduled tasks ---

// ListScheduledTasks lists scheduled tasks for an application.
func (s *ApplicationService) ListScheduledTasks(ctx context.Context, appUUID string) ([]models.ScheduledTask, error) {
	var tasks []models.ScheduledTask
	path := fmt.Sprintf("applications/%s/scheduled-tasks", url.PathEscape(appUUID))
	if err := s.client.Get(ctx, path, &tasks); err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks for application %s: %w", appUUID, err)
	}
	return tasks, nil
}

// CreateScheduledTask creates a scheduled task for an application.
func (s *ApplicationService) CreateScheduledTask(ctx context.Context, appUUID string, req models.ScheduledTaskCreateRequest) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	path := fmt.Sprintf("applications/%s/scheduled-tasks", url.PathEscape(appUUID))
	if err := s.client.Post(ctx, path, req, &task); err != nil {
		return nil, fmt.Errorf("failed to create scheduled task for application %s: %w", appUUID, err)
	}
	return &task, nil
}

// UpdateScheduledTask updates a scheduled task for an application.
func (s *ApplicationService) UpdateScheduledTask(ctx context.Context, appUUID, taskUUID string, req models.ScheduledTaskUpdateRequest) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	path := fmt.Sprintf("applications/%s/scheduled-tasks/%s", url.PathEscape(appUUID), url.PathEscape(taskUUID))
	if err := s.client.Patch(ctx, path, req, &task); err != nil {
		return nil, fmt.Errorf("failed to update scheduled task %s for application %s: %w", taskUUID, appUUID, err)
	}
	return &task, nil
}

// DeleteScheduledTask deletes a scheduled task from an application.
func (s *ApplicationService) DeleteScheduledTask(ctx context.Context, appUUID, taskUUID string) error {
	path := fmt.Sprintf("applications/%s/scheduled-tasks/%s", url.PathEscape(appUUID), url.PathEscape(taskUUID))
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("failed to delete scheduled task %s from application %s: %w", taskUUID, appUUID, err)
	}
	return nil
}

// ListScheduledTaskExecutions lists executions for an application scheduled task.
func (s *ApplicationService) ListScheduledTaskExecutions(ctx context.Context, appUUID, taskUUID string) ([]models.ScheduledTaskExecution, error) {
	var executions []models.ScheduledTaskExecution
	path := fmt.Sprintf("applications/%s/scheduled-tasks/%s/executions", url.PathEscape(appUUID), url.PathEscape(taskUUID))
	if err := s.client.Get(ctx, path, &executions); err != nil {
		return nil, fmt.Errorf("failed to list executions for scheduled task %s on application %s: %w", taskUUID, appUUID, err)
	}
	return executions, nil
}

// ExecuteScheduledTask runs an application scheduled task immediately.
func (s *ApplicationService) ExecuteScheduledTask(ctx context.Context, appUUID, taskUUID string) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("applications/%s/scheduled-tasks/%s/execute", url.PathEscape(appUUID), url.PathEscape(taskUUID))
	if err := s.client.Post(ctx, path, map[string]any{}, &resp); err != nil {
		return nil, fmt.Errorf("failed to execute scheduled task %s on application %s: %w", taskUUID, appUUID, err)
	}
	return resp, nil
}

// --- Service scheduled tasks ---

// ListScheduledTasks lists scheduled tasks for a service.
func (s *Service) ListScheduledTasks(ctx context.Context, serviceUUID string) ([]models.ScheduledTask, error) {
	var tasks []models.ScheduledTask
	path := fmt.Sprintf("services/%s/scheduled-tasks", url.PathEscape(serviceUUID))
	if err := s.client.Get(ctx, path, &tasks); err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks for service %s: %w", serviceUUID, err)
	}
	return tasks, nil
}

// CreateScheduledTask creates a scheduled task for a service.
func (s *Service) CreateScheduledTask(ctx context.Context, serviceUUID string, req models.ScheduledTaskCreateRequest) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	path := fmt.Sprintf("services/%s/scheduled-tasks", url.PathEscape(serviceUUID))
	if err := s.client.Post(ctx, path, req, &task); err != nil {
		return nil, fmt.Errorf("failed to create scheduled task for service %s: %w", serviceUUID, err)
	}
	return &task, nil
}

// UpdateScheduledTask updates a scheduled task for a service.
func (s *Service) UpdateScheduledTask(ctx context.Context, serviceUUID, taskUUID string, req models.ScheduledTaskUpdateRequest) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	path := fmt.Sprintf("services/%s/scheduled-tasks/%s", url.PathEscape(serviceUUID), url.PathEscape(taskUUID))
	if err := s.client.Patch(ctx, path, req, &task); err != nil {
		return nil, fmt.Errorf("failed to update scheduled task %s for service %s: %w", taskUUID, serviceUUID, err)
	}
	return &task, nil
}

// DeleteScheduledTask deletes a scheduled task from a service.
func (s *Service) DeleteScheduledTask(ctx context.Context, serviceUUID, taskUUID string) error {
	path := fmt.Sprintf("services/%s/scheduled-tasks/%s", url.PathEscape(serviceUUID), url.PathEscape(taskUUID))
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("failed to delete scheduled task %s from service %s: %w", taskUUID, serviceUUID, err)
	}
	return nil
}

// ListScheduledTaskExecutions lists executions for a service scheduled task.
func (s *Service) ListScheduledTaskExecutions(ctx context.Context, serviceUUID, taskUUID string) ([]models.ScheduledTaskExecution, error) {
	var executions []models.ScheduledTaskExecution
	path := fmt.Sprintf("services/%s/scheduled-tasks/%s/executions", url.PathEscape(serviceUUID), url.PathEscape(taskUUID))
	if err := s.client.Get(ctx, path, &executions); err != nil {
		return nil, fmt.Errorf("failed to list executions for scheduled task %s on service %s: %w", taskUUID, serviceUUID, err)
	}
	return executions, nil
}

// ExecuteScheduledTask runs a service scheduled task immediately.
func (s *Service) ExecuteScheduledTask(ctx context.Context, serviceUUID, taskUUID string) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("services/%s/scheduled-tasks/%s/execute", url.PathEscape(serviceUUID), url.PathEscape(taskUUID))
	if err := s.client.Post(ctx, path, map[string]any{}, &resp); err != nil {
		return nil, fmt.Errorf("failed to execute scheduled task %s on service %s: %w", taskUUID, serviceUUID, err)
	}
	return resp, nil
}
