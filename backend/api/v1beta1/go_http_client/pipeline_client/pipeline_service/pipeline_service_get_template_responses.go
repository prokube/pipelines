// Code generated by go-swagger; DO NOT EDIT.

package pipeline_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/kubeflow/pipelines/backend/api/v1beta1/go_http_client/pipeline_model"
)

// PipelineServiceGetTemplateReader is a Reader for the PipelineServiceGetTemplate structure.
type PipelineServiceGetTemplateReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PipelineServiceGetTemplateReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPipelineServiceGetTemplateOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewPipelineServiceGetTemplateDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewPipelineServiceGetTemplateOK creates a PipelineServiceGetTemplateOK with default headers values
func NewPipelineServiceGetTemplateOK() *PipelineServiceGetTemplateOK {
	return &PipelineServiceGetTemplateOK{}
}

/*
PipelineServiceGetTemplateOK describes a response with status code 200, with default header values.

A successful response.
*/
type PipelineServiceGetTemplateOK struct {
	Payload *pipeline_model.APIGetTemplateResponse
}

// IsSuccess returns true when this pipeline service get template o k response has a 2xx status code
func (o *PipelineServiceGetTemplateOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this pipeline service get template o k response has a 3xx status code
func (o *PipelineServiceGetTemplateOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this pipeline service get template o k response has a 4xx status code
func (o *PipelineServiceGetTemplateOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this pipeline service get template o k response has a 5xx status code
func (o *PipelineServiceGetTemplateOK) IsServerError() bool {
	return false
}

// IsCode returns true when this pipeline service get template o k response a status code equal to that given
func (o *PipelineServiceGetTemplateOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the pipeline service get template o k response
func (o *PipelineServiceGetTemplateOK) Code() int {
	return 200
}

func (o *PipelineServiceGetTemplateOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /apis/v1beta1/pipelines/{id}/templates][%d] pipelineServiceGetTemplateOK %s", 200, payload)
}

func (o *PipelineServiceGetTemplateOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /apis/v1beta1/pipelines/{id}/templates][%d] pipelineServiceGetTemplateOK %s", 200, payload)
}

func (o *PipelineServiceGetTemplateOK) GetPayload() *pipeline_model.APIGetTemplateResponse {
	return o.Payload
}

func (o *PipelineServiceGetTemplateOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(pipeline_model.APIGetTemplateResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPipelineServiceGetTemplateDefault creates a PipelineServiceGetTemplateDefault with default headers values
func NewPipelineServiceGetTemplateDefault(code int) *PipelineServiceGetTemplateDefault {
	return &PipelineServiceGetTemplateDefault{
		_statusCode: code,
	}
}

/*
PipelineServiceGetTemplateDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type PipelineServiceGetTemplateDefault struct {
	_statusCode int

	Payload *pipeline_model.GooglerpcStatus
}

// IsSuccess returns true when this pipeline service get template default response has a 2xx status code
func (o *PipelineServiceGetTemplateDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this pipeline service get template default response has a 3xx status code
func (o *PipelineServiceGetTemplateDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this pipeline service get template default response has a 4xx status code
func (o *PipelineServiceGetTemplateDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this pipeline service get template default response has a 5xx status code
func (o *PipelineServiceGetTemplateDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this pipeline service get template default response a status code equal to that given
func (o *PipelineServiceGetTemplateDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the pipeline service get template default response
func (o *PipelineServiceGetTemplateDefault) Code() int {
	return o._statusCode
}

func (o *PipelineServiceGetTemplateDefault) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /apis/v1beta1/pipelines/{id}/templates][%d] PipelineService_GetTemplate default %s", o._statusCode, payload)
}

func (o *PipelineServiceGetTemplateDefault) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /apis/v1beta1/pipelines/{id}/templates][%d] PipelineService_GetTemplate default %s", o._statusCode, payload)
}

func (o *PipelineServiceGetTemplateDefault) GetPayload() *pipeline_model.GooglerpcStatus {
	return o.Payload
}

func (o *PipelineServiceGetTemplateDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(pipeline_model.GooglerpcStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
