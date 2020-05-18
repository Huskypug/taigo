package taigo

import (
	"errors"
	"fmt"

	"github.com/google/go-querystring/query"
)

const endpointTasks = "/tasks"

// TaskService is a handle to actions related to Tasks
//
// https://taigaio.github.io/taiga-doc/dist/api.html#tasks
type TaskService struct {
	client *Client
}

// List => https://taigaio.github.io/taiga-doc/dist/api.html#tasks-list
func (s *TaskService) List(queryParameters *TasksQueryParams) ([]Task, error) {
	url := s.client.APIURL + endpointTasks
	if queryParameters != nil {
		paramValues, _ := query.Values(queryParameters)
		url = fmt.Sprintf("%s?%s", url, paramValues.Encode())
	} else if s.client.HasDefaultProject() {
		url = url + s.client.GetDefaultProjectAsQueryParam()
	}

	var taskDetailList TaskDetailLIST

	err := s.client.Request.GetRequest(url, &taskDetailList)
	if err != nil {
		return nil, err
	}
	return taskDetailList.AsTask()
}

// Create creates a new Task | https://taigaio.github.io/taiga-doc/dist/api.html#tasks-create
// Meta Available: *TaskDetail
func (s *TaskService) Create(task *Task) (*Task, error) {
	url := s.client.APIURL + endpointTasks
	var responseTask TaskDetail

	// Check for required fields
	// project, subject
	if isEmpty(task.Project) || isEmpty(task.Subject) {
		return nil, errors.New("A mandatory field is missing. See API documentataion")
	}

	err := s.client.Request.PostRequest(url, &task, &responseTask)
	if err != nil {
		return nil, err
	}
	return responseTask.AsTask()
}

// Get => https://taigaio.github.io/taiga-doc/dist/api.html#tasks-get
func (s *TaskService) Get(task *Task) (*Task, error) {
	url := s.client.APIURL + fmt.Sprintf("%s/%d", endpointTasks, task.ID)
	var responseTask TaskDetailGET
	err := s.client.Request.GetRequest(url, &responseTask)
	if err != nil {
		return nil, err
	}
	return responseTask.AsTask()
}

// GetByRef => https://taigaio.github.io/taiga-doc/dist/api.html#tasks-get-by-ref
func (s *TaskService) GetByRef(task *Task, project *Project) (*Task, error) {
	var respTask TaskDetailGET
	var url string
	if project.ID != 0 {
		url = s.client.APIURL + fmt.Sprintf("%s/by_ref?ref=%d&project=%d", endpointTasks, task.Ref, project.ID)
	} else if len(project.Slug) > 0 {
		url = s.client.APIURL + fmt.Sprintf("%s/by_ref?ref=%d&project__slug=%s", endpointTasks, task.Ref, project.Slug)
	} else {
		return nil, errors.New("No ID or Ref defined in passed project struct")
	}

	err := s.client.Request.GetRequest(url, &respTask)
	if err != nil {
		return nil, err
	}
	return respTask.AsTask()
}

// GetAttachment retrives a Task attachment by its ID => https://taigaio.github.io/taiga-doc/dist/api.html#tasks-get-attachment
func (s *TaskService) GetAttachment(attachment *Attachment) (*Attachment, error) {
	a, err := getAttachmentForEndpoint(s.client, attachment, endpointTasks)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ListAttachments returns a list of Task attachments => https://taigaio.github.io/taiga-doc/dist/api.html#tasks-list-attachments
func (s *TaskService) ListAttachments(task interface{}) (*[]Attachment, error) {
	t := Task{}
	err := convertStructViaJSON(task, &t)
	if err != nil {
		return nil, err
	}

	queryParams := attachmentsQueryParams{
		endpointURI: endpointTasks,
		ObjectID:    t.ID,
		Project:     t.Project,
	}

	attachmentsOfTask, err := listAttachmentsForEndpoint(s.client, &queryParams)
	if err != nil {
		return nil, err
	}
	return attachmentsOfTask, nil
}

// CreateAttachment creates a new Task attachment => https://taigaio.github.io/taiga-doc/dist/api.html#tasks-create-attachment
func (s *TaskService) CreateAttachment(attachment *Attachment, task *Task, filePath string) (*Attachment, error) {
	url := s.client.APIURL + endpointTasks + "/attachments"
	attachment.filePath = filePath
	attachment.ObjectID = task.ID
	if attachment.Project == 0 && task.Project > 0 {
		attachment.Project = task.Project
	} else {
		return nil, fmt.Errorf("Project.ID could not be fetched from any possible sources")
	}
	return newfileUploadRequest(s.client, url, attachment)
}
