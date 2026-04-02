package db

import (
	"log"

	"treehole/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

// NewDatabase 创建新的数据库实例
func NewDatabase(dbPath string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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

// initTables 初始化数据库表
func (d *Database) initTables() error {
	return d.db.AutoMigrate(&models.Post{}, &models.Comment{})
}

// UpsertPosts 插入或更新帖子
func (d *Database) UpsertPosts(posts []models.Post) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		for _, post := range posts {
			if err := tx.Where("pid = ?", post.Pid).Assign(models.Post{
				Text:       post.Text,
				Anonymous:  post.Anonymous,
				Type:       post.Type,
				ImageSizeX: post.ImageSizeX,
				ImageSizeY: post.ImageSizeY,
				Extra:      post.Extra,
				Timestamp:  post.Timestamp,
				Reply:      post.Reply,
				Likenum:    post.Likenum,
				Tag:        post.Tag,
				Status:     post.Status,
				IsComment:  post.IsComment,
				IsProtect:  post.IsProtect,
				IsTop:      post.IsTop,
				Label:      post.Label,
				MediaIds:   post.MediaIds,
			}).FirstOrCreate(&post).Error; err != nil {
				log.Printf("Error upserting post %d: %v", post.Pid, err)
				return err
			}
		}
		return nil
	})
}

// UpsertComments 插入或更新评论
func (d *Database) UpsertComments(comments []models.Comment) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		for _, comment := range comments {
			if err := tx.Where("cid = ?", comment.Cid).Assign(models.Comment{
				Pid:       comment.Pid,
				Name:      comment.Name,
				Text:      comment.Text,
				Timestamp: comment.Timestamp,
				Tag:       comment.Tag,
				Quote:     comment.Quote,
				MediaIds:  comment.MediaIds,
			}).FirstOrCreate(&comment).Error; err != nil {
				log.Printf("Error upserting comment %d: %v", comment.Cid, err)
				return err
			}
		}
		return nil
	})
}

// Close 关闭数据库连接
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

// GetPostCount 获取帖子总数
func (d *Database) GetPostCount() (int, error) {
	var count int64
	err := d.db.Model(&models.Post{}).Count(&count).Error
	return int(count), err
}

// GetCommentCount 获取评论总数
func (d *Database) GetCommentCount() (int, error) {
	var count int64
	err := d.db.Model(&models.Comment{}).Count(&count).Error
	return int(count), err
}

// GetPosts 分页获取帖子列表
func (d *Database) GetPosts(offset, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := d.db.Order("pid DESC").Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

// GetPostByPid 根据pid获取帖子
func (d *Database) GetPostByPid(pid int) (*models.Post, error) {
	var post models.Post
	err := d.db.Where("pid = ?", pid).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetCommentsByPid 根据pid获取评论列表
func (d *Database) GetCommentsByPid(pid int) ([]models.Comment, error) {
	var comments []models.Comment
	err := d.db.Where("pid = ?", pid).Order("cid ASC").Find(&comments).Error
	return comments, err
}

// SearchPosts 搜索帖子
func (d *Database) SearchPosts(keyword string, offset, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := d.db.Where("text LIKE ?", "%"+keyword+"%").Order("pid DESC").Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

// SearchPostsCount 搜索帖子总数
func (d *Database) SearchPostsCount(keyword string) (int, error) {
	var count int64
	err := d.db.Model(&models.Post{}).Where("text LIKE ?", "%"+keyword+"%").Count(&count).Error
	return int(count), err
}
