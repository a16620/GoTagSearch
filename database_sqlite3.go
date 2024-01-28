package model

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteHandler struct {
	db  *sql.DB
	rwm *sync.RWMutex
}

func checkError(err error) {
	if err != nil {
		log.Default().Println(err.Error())
		panic(err)
	}
}

func NewSqlite(file string) *SqliteHandler {
	db, err := sql.Open("sqlite3", file)
	checkError(err)

	if err != nil {
		return nil
	} else {
		_sql := &SqliteHandler{
			db:  db,
			rwm: new(sync.RWMutex),
		}
		return _sql
	}
}

func (sh *SqliteHandler) Query(query string) error {
	_, err := sh.db.Exec(query)
	return err
}

func (sh *SqliteHandler) Close() {
	sh.rwm.Lock()
	sh.db.Close()
	sh.rwm.Unlock()
}

func (sh *SqliteHandler) Init() {
	query := `
		CREATE TABLE IF NOT EXISTS article (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT UNIQUE NOT NULL,
			platform TEXT NOT NULL,
			description TEXT,
			thumbnail_url TEXT
		);
	`
	checkError(sh.Query(query))

	query = `
		CREATE TABLE IF NOT EXISTS tag (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type INTEGER DEFAULT 0,
			UNIQUE(name, type)
		);
	`
	checkError(sh.Query(query))

	query = `
		CREATE TABLE IF NOT EXISTS article_tag (
			article_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			UNIQUE(article_id, tag_id)
		);
	`
	checkError(sh.Query(query))

	query = `
		CREATE INDEX IF NOT EXISTS tag_name_index on tag(name);
	`
	checkError(sh.Query(query))
}

func (sh *SqliteHandler) GetArticles() []*Article {
	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query("SELECT * from article")
	checkError(err)
	defer result.Close()

	articles := []*Article{}
	for result.Next() {
		var art Article
		err = result.Scan(&art.ID, &art.Url, &art.Platform, &art.Description, &art.Thumbnail_url)
		checkError(err)

		articles = append(articles, &art)
	}

	return articles
}

func (sh *SqliteHandler) GetArticlesByTags(tags []*Tag) []*Article {
	//article_tag에 tag join, HAVING COUNT(*) >= len(tags)
	tag_count := len(tags)
	const placeholder = ",(?,?)"
	query := `
	SELECT article.* FROM
	article INNER JOIN
	(SELECT article_tag.article_id as id FROM article_tag INNER JOIN tag ON article_tag.tag_id = tag.id WHERE (tag.name, tag.type) IN (VALUES (?,?)` + strings.Repeat(placeholder, tag_count-1) + `) GROUP BY article_tag.article_id HAVING count(*) >= ?) subq
	ON article.id = subq.id
	`

	args := make([]interface{}, tag_count*2+1)
	for i, v := range tags {
		args[2*i] = v.Name
		args[2*i+1] = v.Type
	}
	args[tag_count*2] = tag_count

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query, args...)
	checkError(err)
	defer result.Close()

	articles := []*Article{}
	for result.Next() {
		var art Article
		err = result.Scan(&art.ID, &art.Url, &art.Platform, &art.Description, &art.Thumbnail_url)
		checkError(err)

		articles = append(articles, &art)
	}

	return articles
}

func (sh *SqliteHandler) GetArticlesByTagName(tags []string) []*Article {
	tag_count := len(tags)
	query := `
	SELECT article.* FROM
	article INNER JOIN
	(SELECT article_tag.article_id as id FROM article_tag INNER JOIN tag ON article_tag.tag_id = tag.id WHERE tag.name IN (?` + strings.Repeat(",?", tag_count-1) + `) GROUP BY article_tag.article_id HAVING count(*) >= ?) subq
	ON article.id = subq.id
	`

	args := make([]interface{}, tag_count+1)
	for i, v := range tags {
		args[i] = v
	}
	args[tag_count] = tag_count

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query, args...)
	checkError(err)
	defer result.Close()

	articles := []*Article{}
	for result.Next() {
		var art Article
		err = result.Scan(&art.ID, &art.Url, &art.Platform, &art.Description, &art.Thumbnail_url)
		checkError(err)

		articles = append(articles, &art)
	}

	return articles
}

func (sh *SqliteHandler) GetArticlesByTagID(tags []int) []*Article {
	tag_count := len(tags)
	query := `
	SELECT article.* FROM
	article INNER JOIN
	(SELECT article_id as id FROM article_tag WHERE tag_id IN (?` + strings.Repeat(",?", tag_count-1) + `) GROUP BY article_id HAVING count(*) >= ?) subq
	ON article.id = subq.id
	`

	args := make([]interface{}, tag_count+1)
	for i, v := range tags {
		args[i] = v
	}
	args[tag_count] = tag_count

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query, args...)
	checkError(err)
	defer result.Close()

	articles := []*Article{}
	for result.Next() {
		var art Article
		err = result.Scan(&art.ID, &art.Url, &art.Platform, &art.Description, &art.Thumbnail_url)
		checkError(err)

		articles = append(articles, &art)
	}

	return articles
}

func (sh *SqliteHandler) GetTagList() []*Tag {
	query := `
		SELECT * from tag;
	`

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query)
	checkError(err)
	defer result.Close()

	tags := []*Tag{}
	for result.Next() {
		var t Tag
		err = result.Scan(&t.ID, &t.Name, &t.Type)
		checkError(err)

		tags = append(tags, &t)
	}

	return tags
}

func (sh *SqliteHandler) GetTagsContaining(substr string) []*Tag {
	//Escape substr?
	query := `
		SELECT * from tag WHERE name LIKE '%` + substr + `%'
	`

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query)
	checkError(err)
	defer result.Close()

	tags := []*Tag{}
	for result.Next() {
		var t Tag
		err = result.Scan(&t.ID, &t.Name, &t.Type)
		checkError(err)

		tags = append(tags, &t)
	}

	return tags
}

func (sh *SqliteHandler) GetTagOfArticle(article_id int) []*Tag {
	query := `
		SELECT tag.* FROM article_tag JOIN tag ON article_tag.tag_id = tag.id WHERE article_tag.article_id = ?;
	`

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query, article_id)
	checkError(err)
	defer result.Close()

	tags := []*Tag{}
	for result.Next() {
		var t Tag
		err = result.Scan(&t.ID, &t.Name, &t.Type)
		checkError(err)

		tags = append(tags, &t)
	}

	return tags
}

func (sh *SqliteHandler) AddArticle(ainfo *Article) error {
	sh.rwm.Lock()
	defer sh.rwm.Unlock()

	_, err := sh.db.Exec("INSERT INTO article (url, platform, description, thumbnail_url) VALUES (?,?,?,?);", ainfo.Url, ainfo.Platform, ainfo.Description, ainfo.Thumbnail_url)
	return err
}

func (sh *SqliteHandler) AddTags(tags []*Tag) error {
	var query_buf strings.Builder
	query_buf.WriteString("INSERT OR IGNORE INTO tag (name, type) VALUES ")

	for _, v := range tags {
		query_buf.WriteString("('")
		query_buf.WriteString(v.Name)
		query_buf.WriteString("',")
		query_buf.WriteString(strconv.Itoa(v.Type))
		query_buf.WriteString("),")
	}

	query := query_buf.String()
	query = query[:len(query)-1] + ";"

	sh.rwm.Lock()
	defer sh.rwm.Unlock()

	return sh.Query(query)
}

func (sh *SqliteHandler) mappingTagID(tags []*Tag) []*Tag {
	mapped_tags := []*Tag{}
	unmapped_tags := []*Tag{}
	for _, v := range tags {
		if v.ID == 0 {
			unmapped_tags = append(unmapped_tags, v)
		} else {
			mapped_tags = append(mapped_tags, v)
		}
	}

	tag_count, total := len(unmapped_tags), len(tags)

	if tag_count == 0 {
		return tags
	}

	const placeholder = ",(?,?)"
	query := `
		SELECT * FROM tag WHERE (name, type) IN (VALUES (?,?) ` + strings.Repeat(placeholder, tag_count-1) + `)
	`

	args := make([]interface{}, tag_count*2)
	for i, v := range unmapped_tags {
		args[2*i] = v.Name
		args[2*i+1] = v.Type
	}

	sh.rwm.RLock()
	defer sh.rwm.RUnlock()

	result, err := sh.db.Query(query, args...)
	checkError(err)
	defer result.Close()

	i := 0
	for result.Next() {
		ptr := unmapped_tags[i]
		err = result.Scan(&ptr.ID, &ptr.Name, &ptr.Type)
		checkError(err)

		tags[i] = ptr
		i += 1
	}

	if i != tag_count {
		panic("tag mapping fail")
	}

	for i < total {
		tags[i] = mapped_tags[i-tag_count]
		i += 1
	}

	return tags
}

func (sh *SqliteHandler) AttachTagsToArticle(article_id int, tags []*Tag) error {
	sh.AddTags(tags)
	tags = sh.mappingTagID(tags)

	var query_buf strings.Builder
	query_buf.WriteString("INSERT OR IGNORE INTO article_tag (article_id, tag_id) VALUES ")

	beg := "(" + strconv.Itoa(article_id) + ","

	for _, v := range tags {
		query_buf.WriteString(beg)
		query_buf.WriteString(strconv.Itoa(v.ID))
		query_buf.WriteString("),")
	}

	query := query_buf.String()
	query = query[:len(query)-1] + ";"

	sh.rwm.Lock()
	defer sh.rwm.Unlock()

	return sh.Query(query)
}
