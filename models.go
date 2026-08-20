package main

import "encoding/json"

// =============================================================================
// Generic Webhook Payload
// =============================================================================

// WebhookPayload is the top-level structure sent by OpenProject webhooks.
// The "action" field indicates the event type (e.g., "work_package:created").
// The actual resource data is present in a field named after the resource type.
type WebhookPayload struct {
	Action      string          `json:"action"`
	WorkPackage json.RawMessage `json:"work_package,omitempty"`
}

// =============================================================================
// HAL Link
// =============================================================================

// HALLink represents a single HAL+JSON link with href and title.
type HALLink struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

// =============================================================================
// Work Package
// =============================================================================

// WorkPackageLinks holds the _links section of a work package resource.
type WorkPackageLinks struct {
	Self        HALLink `json:"self"`
	Project     HALLink `json:"project"`
	Type        HALLink `json:"type"`
	Status      HALLink `json:"status"`
	Priority    HALLink `json:"priority"`
	Author      HALLink `json:"author"`
	Assignee    HALLink `json:"assignee"`
	Responsible HALLink `json:"responsible"`
	Version     HALLink `json:"version"`
}

// TextNode represents a rich-text field in OpenProject (description, etc.).
type TextNode struct {
	Format string `json:"format"`
	Raw    string `json:"raw"`
	HTML   string `json:"html"`
}

// WorkPackage represents the work package resource from the webhook payload.
type WorkPackage struct {
	Type        string           `json:"_type"`
	ID          int              `json:"id"`
	LockVersion int              `json:"lockVersion"`
	Subject     string           `json:"subject"`
	Description TextNode         `json:"description"`
	CreatedAt   string           `json:"createdAt"`
	UpdatedAt   string           `json:"updatedAt"`
	Links       WorkPackageLinks `json:"_links"`
}
