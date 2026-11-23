package udemyapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const curriculumPathTemplate = "/api-2.0/courses/%d/subscriber-curriculum-items/?"

var curriculumParams = url.Values{
	"fields[lecture]":  {"title,object_index,created,asset,supplementary_assets,description,download_url"},
	"fields[quiz]":     {"title,object_index,type"},
	"fields[practice]": {"title,object_index"},
	"fields[chapter]":  {"title,object_index"},
	"fields[asset]":    {"title,filename,asset_type,status,is_external,media_license_token,course_is_drmed,media_sources,captions,slides,slide_urls,download_urls,external_url,stream_urls,@min,status,delayed_asset_message,processing_errors,body"},
	"caching_intent":   {"True"},
	"page_size":        {"200"},
}

func (client *Client) Curriculum(ctx context.Context, courseID int64) ([]udemy.CurriculumItem, error) {
	target := client.portalURL + fmt.Sprintf(curriculumPathTemplate, courseID) + curriculumParams.Encode()
	items, err := allPages[udemy.CurriculumItem](ctx, client, target)
	if err != nil {
		return nil, fmt.Errorf("fetch curriculum of course %d: %w", courseID, err)
	}
	return items, nil
}
