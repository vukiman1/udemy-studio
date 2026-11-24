package udemyapi

import (
	"context"
	"fmt"
)

const courseCoverPathTemplate = "/api-2.0/courses/%d/?fields%%5Bcourse%%5D=image_750x422"

// CourseCoverURL is the course's 750x422 cover image on Udemy's CDN, empty when the course has none.
func (client *Client) CourseCoverURL(ctx context.Context, courseID int64) (string, error) {
	var course struct {
		Image string `json:"image_750x422"`
	}
	if err := client.getJSON(ctx, client.portalURL+fmt.Sprintf(courseCoverPathTemplate, courseID), &course); err != nil {
		return "", fmt.Errorf("fetch the cover of course %d: %w", courseID, err)
	}
	return course.Image, nil
}
