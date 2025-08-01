// Code generated by go-swagger; DO NOT EDIT.

package run_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewRunServiceArchiveRunParams creates a new RunServiceArchiveRunParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewRunServiceArchiveRunParams() *RunServiceArchiveRunParams {
	return &RunServiceArchiveRunParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewRunServiceArchiveRunParamsWithTimeout creates a new RunServiceArchiveRunParams object
// with the ability to set a timeout on a request.
func NewRunServiceArchiveRunParamsWithTimeout(timeout time.Duration) *RunServiceArchiveRunParams {
	return &RunServiceArchiveRunParams{
		timeout: timeout,
	}
}

// NewRunServiceArchiveRunParamsWithContext creates a new RunServiceArchiveRunParams object
// with the ability to set a context for a request.
func NewRunServiceArchiveRunParamsWithContext(ctx context.Context) *RunServiceArchiveRunParams {
	return &RunServiceArchiveRunParams{
		Context: ctx,
	}
}

// NewRunServiceArchiveRunParamsWithHTTPClient creates a new RunServiceArchiveRunParams object
// with the ability to set a custom HTTPClient for a request.
func NewRunServiceArchiveRunParamsWithHTTPClient(client *http.Client) *RunServiceArchiveRunParams {
	return &RunServiceArchiveRunParams{
		HTTPClient: client,
	}
}

/*
RunServiceArchiveRunParams contains all the parameters to send to the API endpoint

	for the run service archive run operation.

	Typically these are written to a http.Request.
*/
type RunServiceArchiveRunParams struct {

	/* ExperimentID.

	   The ID of the parent experiment.
	*/
	ExperimentID *string

	/* RunID.

	   The ID of the run to be archived.
	*/
	RunID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the run service archive run params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *RunServiceArchiveRunParams) WithDefaults() *RunServiceArchiveRunParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the run service archive run params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *RunServiceArchiveRunParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the run service archive run params
func (o *RunServiceArchiveRunParams) WithTimeout(timeout time.Duration) *RunServiceArchiveRunParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the run service archive run params
func (o *RunServiceArchiveRunParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the run service archive run params
func (o *RunServiceArchiveRunParams) WithContext(ctx context.Context) *RunServiceArchiveRunParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the run service archive run params
func (o *RunServiceArchiveRunParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the run service archive run params
func (o *RunServiceArchiveRunParams) WithHTTPClient(client *http.Client) *RunServiceArchiveRunParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the run service archive run params
func (o *RunServiceArchiveRunParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithExperimentID adds the experimentID to the run service archive run params
func (o *RunServiceArchiveRunParams) WithExperimentID(experimentID *string) *RunServiceArchiveRunParams {
	o.SetExperimentID(experimentID)
	return o
}

// SetExperimentID adds the experimentId to the run service archive run params
func (o *RunServiceArchiveRunParams) SetExperimentID(experimentID *string) {
	o.ExperimentID = experimentID
}

// WithRunID adds the runID to the run service archive run params
func (o *RunServiceArchiveRunParams) WithRunID(runID string) *RunServiceArchiveRunParams {
	o.SetRunID(runID)
	return o
}

// SetRunID adds the runId to the run service archive run params
func (o *RunServiceArchiveRunParams) SetRunID(runID string) {
	o.RunID = runID
}

// WriteToRequest writes these params to a swagger request
func (o *RunServiceArchiveRunParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.ExperimentID != nil {

		// query param experiment_id
		var qrExperimentID string

		if o.ExperimentID != nil {
			qrExperimentID = *o.ExperimentID
		}
		qExperimentID := qrExperimentID
		if qExperimentID != "" {

			if err := r.SetQueryParam("experiment_id", qExperimentID); err != nil {
				return err
			}
		}
	}

	// path param run_id
	if err := r.SetPathParam("run_id", o.RunID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
