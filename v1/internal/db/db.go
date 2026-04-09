package db

import (
	"fmt"
	"log"
	"strings"
	"treehole/internal/config"
	"treehole/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// containsNullByte checks if a string contains null bytes (0x00)
func containsNullByte(s string) bool {
	return strings.ContainsRune(s, '\x00')
}

// sanitizeNullBytes removes null bytes from a string
func sanitizeNullBytes(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

// escapeLikePattern escapes % and _ characters in a LIKE pattern to prevent wildcard injection
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return "%" + s + "%"
}

type Database struct {
	db *gorm.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn, err := cfg.GetDatabaseDSN()
	if err != nil {
		return nil, err
	}

	var db *gorm.DB
	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	switch cfg.Database.Type {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(dsn), gormCfg)
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	database := &Database{db: db}
	err = database.initTables()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) initTables() error {
	return d.db.AutoMigrate(&models.Post{}, &models.Comment{}, &models.ExclusiveIdInfo{})
}

func (d *Database) UpsertPosts(posts []models.Post) error {
	// Check and sanitize null bytes before writing
	for i := range posts {
		if containsNullByte(posts[i].Text) {
			log.Printf("[Database] Post pid=%d 包含 null 字节，已清理。Text前100字符: %q",
				posts[i].Pid, sanitizeNullBytes(posts[i].Text[:min(len(posts[i].Text), 100)]))
			posts[i].Text = sanitizeNullBytes(posts[i].Text)
		}
		if containsNullByte(posts[i].IdentityInfo) {
			log.Printf("[Database] Post pid=%d IdentityInfo 包含 null 字节，已清理", posts[i].Pid)
			posts[i].IdentityInfo = sanitizeNullBytes(posts[i].IdentityInfo)
		}
		if containsNullByte(posts[i].ExclusiveIdInfo) {
			log.Printf("[Database] Post pid=%d ExclusiveIdInfo 包含 null 字节，已清理", posts[i].Pid)
			posts[i].ExclusiveIdInfo = sanitizeNullBytes(posts[i].ExclusiveIdInfo)
		}
		if containsNullByte(posts[i].Mention) {
			log.Printf("[Database] Post pid=%d Mention 包含 null 字节，已清理", posts[i].Pid)
			posts[i].Mention = sanitizeNullBytes(posts[i].Mention)
		}
		if containsNullByte(posts[i].ImageSize) {
			posts[i].ImageSize = sanitizeNullBytes(posts[i].ImageSize)
		}

		// 检查嵌套的评论
		for j := range posts[i].Comments {
			if containsNullByte(posts[i].Comments[j].Text) {
				log.Printf("[Database] Post pid=%d 的评论 cid=%d 包含 null 字节，已清理。Text前100字符: %q",
					posts[i].Pid, posts[i].Comments[j].Cid, sanitizeNullBytes(posts[i].Comments[j].Text[:min(len(posts[i].Comments[j].Text), 100)]))
				posts[i].Comments[j].Text = sanitizeNullBytes(posts[i].Comments[j].Text)
			}
			if containsNullByte(posts[i].Comments[j].IdentityInfo) {
				log.Printf("[Database] Post pid=%d 的评论 cid=%d IdentityInfo 包含 null 字节，已清理", posts[i].Pid, posts[i].Comments[j].Cid)
				posts[i].Comments[j].IdentityInfo = sanitizeNullBytes(posts[i].Comments[j].IdentityInfo)
			}
			if containsNullByte(posts[i].Comments[j].ExclusiveIdInfo) {
				log.Printf("[Database] Post pid=%d 的评论 cid=%d ExclusiveIdInfo 包含 null 字节，已清理", posts[i].Pid, posts[i].Comments[j].Cid)
				posts[i].Comments[j].ExclusiveIdInfo = sanitizeNullBytes(posts[i].Comments[j].ExclusiveIdInfo)
			}
			if containsNullByte(posts[i].Comments[j].Mention) {
				posts[i].Comments[j].Mention = sanitizeNullBytes(posts[i].Comments[j].Mention)
			}
			if containsNullByte(posts[i].Comments[j].NameTag) {
				posts[i].Comments[j].NameTag = sanitizeNullBytes(posts[i].Comments[j].NameTag)
			}
		}
	}

	err := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pid"}},
		UpdateAll: true,
	}).CreateInBatches(posts, 100).Error

	if err != nil {
		pids := make([]int32, len(posts))
		for i, p := range posts {
			pids[i] = p.Pid
		}
		log.Printf("[Database] UpsertPosts 失败: %v | 涉及帖子 PIDs: %v", err, pids)
	}

	return err
}

func (d *Database) UpsertComments(comments []models.Comment) error {
	// Check and sanitize null bytes before writing
	for i := range comments {
		if containsNullByte(comments[i].Text) {
			log.Printf("[Database] Comment cid=%d pid=%d 包含 null 字节，已清理。Text前100字符: %q",
				comments[i].Cid, comments[i].Pid, sanitizeNullBytes(comments[i].Text[:min(len(comments[i].Text), 100)]))
			comments[i].Text = sanitizeNullBytes(comments[i].Text)
		}
		if containsNullByte(comments[i].IdentityInfo) {
			log.Printf("[Database] Comment cid=%d IdentityInfo 包含 null 字节，已清理", comments[i].Cid)
			comments[i].IdentityInfo = sanitizeNullBytes(comments[i].IdentityInfo)
		}
		if containsNullByte(comments[i].ExclusiveIdInfo) {
			log.Printf("[Database] Comment cid=%d ExclusiveIdInfo 包含 null 字节，已清理", comments[i].Cid)
			comments[i].ExclusiveIdInfo = sanitizeNullBytes(comments[i].ExclusiveIdInfo)
		}
		if containsNullByte(comments[i].Mention) {
			comments[i].Mention = sanitizeNullBytes(comments[i].Mention)
		}
		if containsNullByte(comments[i].NameTag) {
			comments[i].NameTag = sanitizeNullBytes(comments[i].NameTag)
		}
	}

	err := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cid"}},
		UpdateAll: true,
	}).CreateInBatches(comments, 100).Error

	if err != nil {
		cids := make([]int32, len(comments))
		for i, c := range comments {
			cids[i] = c.Cid
		}
		log.Printf("[Database] UpsertComments 失败: %v | 涉及评论 CIDs: %v", err, cids)
	}

	return err
}

func (d *Database) UpsertExclusiveIdInfo(info models.ExclusiveIdInfo) error {
	return d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&info).Error
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Checkpoint 执行WAL checkpoint
func (d *Database) Checkpoint() error {
	return d.db.Exec("PRAGMA wal_checkpoint(RESTART)").Error
}

func (d *Database) GetPostCount() (int, error) {
	var count int64
	err := d.db.Model(&models.Post{}).Count(&count).Error
	return int(count), err
}

func (d *Database) GetCommentCount() (int, error) {
	var count int64
	err := d.db.Model(&models.Comment{}).Count(&count).Error
	return int(count), err
}

func (d *Database) GetPostByPid(pid int32) (*models.Post, error) {
	var post models.Post
	err := d.db.Raw("SELECT pid, text, anonymous, type, extra, timestamp, reply, likenum, status, is_comment, is_protect, is_top, label, media_ids FROM posts WHERE pid = ?", pid).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPostsCursor 游标分页获取帖子列表 (DESC)
func (d *Database) GetPostsCursor(cursor int, limit int, sortAsc bool) ([]models.Post, error) {
	var posts []models.Post
	order := "DESC"
	comparator := "<"
	if sortAsc {
		order = "ASC"
		comparator = ">"
	}

	if cursor != 0 {
		err := d.db.Raw("SELECT pid, text, anonymous, type, extra, timestamp, reply, likenum, status, is_comment, is_protect, is_top, label, media_ids FROM posts WHERE pid "+comparator+" ? ORDER BY pid "+order+" LIMIT ?", cursor, limit).Scan(&posts).Error
		return posts, err
	}
	err := d.db.Raw("SELECT pid, text, anonymous, type, extra, timestamp, reply, likenum, status, is_comment, is_protect, is_top, label, media_ids FROM posts ORDER BY pid "+order+" LIMIT ?", limit).Scan(&posts).Error
	return posts, err
}

func (d *Database) SearchPostsCursor(keyword string, cursor int, limit int, sortAsc bool) ([]models.Post, error) {
	var posts []models.Post
	order := "DESC"
	comparator := "<"
	if sortAsc {
		order = "ASC"
		comparator = ">"
	}

	if cursor != 0 {
		err := d.db.Raw("SELECT pid, text, anonymous, type, extra, timestamp, reply, likenum, status, is_comment, is_protect, is_top, label, media_ids FROM posts WHERE text LIKE ? AND pid "+comparator+" ? ORDER BY pid "+order+" LIMIT ?", escapeLikePattern(keyword), cursor, limit).Scan(&posts).Error
		return posts, err
	}
	err := d.db.Raw("SELECT pid, text, anonymous, type, extra, timestamp, reply, likenum, status, is_comment, is_protect, is_top, label, media_ids FROM posts WHERE text LIKE ? ORDER BY pid "+order+" LIMIT ?", escapeLikePattern(keyword), limit).Scan(&posts).Error
	return posts, err
}

// GetCommentsByPidCursor 游标分页获取评论列表
func (d *Database) GetCommentsByPidCursor(pid int32, cursor int32, limit int, sortAsc bool) ([]models.Comment, error) {
	var comments []models.Comment
	query := d.db.Model(&models.Comment{}).Preload("Quote").Where("pid = ?", pid)

	if sortAsc {
		query = query.Order("cid ASC")
		if cursor != 0 {
			query = query.Where("cid > ?", cursor)
		}
	} else {
		query = query.Order("cid DESC")
		if cursor != 0 {
			query = query.Where("cid < ?", cursor)
		}
	}

	err := query.Limit(limit).Find(&comments).Error
	return comments, err
}

func (d *Database) GetPostsOrderBy(field string, cursor int, limit int) ([]models.Post, error) {
	var posts []models.Post
	orderCol := validateOrderField(field)

	query := d.db.Model(&models.Post{}).Order(orderCol + " DESC")
	if cursor != 0 {
		query = query.Where(orderCol+" < ?", cursor)
	}

	err := query.Limit(limit).Find(&posts).Error
	return posts, err
}

func (d *Database) SearchPostsOrderBy(keyword string, field string, cursor int, limit int) ([]models.Post, error) {
	var posts []models.Post
	orderCol := validateOrderField(field)

	query := d.db.Model(&models.Post{}).Where("text LIKE ?", escapeLikePattern(keyword)).Order(orderCol + " DESC")
	if cursor != 0 {
		query = query.Where(orderCol+" < ?", cursor)
	}

	err := query.Limit(limit).Find(&posts).Error
	return posts, err
}

// validateOrderField returns a safe column name from user input
func validateOrderField(field string) string {
	switch field {
	case "reply", "likenum", "praise_num":
		return field
	default:
		return "pid"
	}
}

// GetPostsWithImages 获取有图片的帖子（type=image 或 media_ids 不为空）
func (d *Database) GetPostsWithImages(offset, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := d.db.Model(&models.Post{}).
		Select("pid, type, media_ids").
		Where("type = ? OR media_ids != ?", "image", "").
		Order("pid DESC").
		Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (d *Database) GetPostsWithImagesCount() (int, error) {
	var count int64
	err := d.db.Model(&models.Post{}).Where("type = ? OR media_ids != ?", "image", "").Count(&count).Error
	return int(count), err
}

// GetCommentsWithImages 获取有图片的评论（media_ids 不为空）
func (d *Database) GetCommentsWithImages(offset, limit int) ([]models.Comment, error) {
	var comments []models.Comment
	err := d.db.Model(&models.Comment{}).
		Select("cid, pid, media_ids").
		Where("media_ids != ?", "").
		Order("cid DESC").
		Offset(offset).Limit(limit).Find(&comments).Error
	return comments, err
}

func (d *Database) GetCommentsWithImagesCount() (int, error) {
	var count int64
	err := d.db.Model(&models.Comment{}).Where("media_ids != ?", "").Count(&count).Error
	return int(count), err
}
