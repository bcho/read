package publisher

import (
	"github.com/bcho/timespan"
	medium "github.com/medium/medium-sdk-go"
)

type mediumPublisher struct {
	client *medium.Medium
	user   *medium.User
}

func NewMediumPublisher(token string) *mediumPublisher {
	return &mediumPublisher{client: medium.NewClientWithAccessToken(token)}
}

func (m mediumPublisher) Publish(span timespan.Span, articles []string) (string, error) {
	var err error

	user, err := m.client.GetUser()
	if err != nil {
		return "", err
	}

	post, err := m.client.CreatePost(medium.CreatePostOptions{
		UserID:        user.ID,
		Title:         title(span, articles),
		Content:       content(span, articles),
		ContentFormat: medium.ContentFormatHTML,
		PublishStatus: medium.PublishStatusDraft,
	})
	if err != nil {
		return "", err
	}

	return post.URL, nil
}
