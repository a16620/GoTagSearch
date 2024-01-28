package model

type Article struct {
	ID            int            `json:"id"`
	Url           string         `json:"url"`
	Platform      string         `json:"platform"`
	Description   NullableString `json:"description"`
	Thumbnail_url NullableString `json:"thumbnail_url"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

type DBHandler interface {
	Init()
	Close()

	GetArticles() []*Article
	AddArticle(ainfo *Article) error
	GetArticlesByTags(tags []*Tag) []*Article
	GetArticlesByTagName(tags []string) []*Article
	GetArticlesByTagID(tags []int) []*Article

	GetTagList() []*Tag
	AddTags(tags []*Tag) error
	GetTagsContaining(substr string) []*Tag

	GetTagOfArticle(article_id int) []*Tag
	AttachTagsToArticle(article_id int, tags []*Tag) error
}

func (tag *Tag) Print() {
	println(tag.ID, tag.Name, tag.Type)
}

func (app *Article) Print() {
	println(app.ID, app.Url)
}
