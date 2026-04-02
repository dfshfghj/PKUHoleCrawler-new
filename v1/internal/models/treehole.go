package models

// Post 表示帖子数据结构
type Post struct {
	Pid        int    `json:"pid" gorm:"primaryKey;column:pid"`
	Text       string `json:"text" gorm:"column:text"`
	Anonymous  int    `json:"anonymous" gorm:"column:anonymous"`
	Type       string `json:"type" gorm:"column:type"`
	ImageSizeX int    `json:"image_size_x" gorm:"column:image_size_x"`
	ImageSizeY int    `json:"image_size_y" gorm:"column:image_size_y"`
	Extra      int    `json:"extra" gorm:"column:extra"`
	Timestamp  int    `json:"timestamp" gorm:"column:timestamp"`
	Reply      int    `json:"reply" gorm:"column:reply"`
	Likenum    int    `json:"likenum" gorm:"column:likenum"`
	Tag        string `json:"tag" gorm:"column:tag"`
	Status     int    `json:"status" gorm:"column:status"`
	IsComment  int    `json:"is_comment" gorm:"column:is_comment"`
	IsProtect  int    `json:"is_protect" gorm:"column:is_protect"`
	IsTop      int    `json:"is_top" gorm:"column:is_top"`
	Label      int    `json:"label" gorm:"column:label"`
	MediaIds   string `json:"media_ids" gorm:"column:media_ids"`
}

func (Post) TableName() string {
	return "posts"
}

// Comment 表示评论数据结构
type Comment struct {
	Cid       int    `json:"cid" gorm:"primaryKey;column:cid"`
	Pid       int    `json:"pid" gorm:"column:pid"`
	Name      string `json:"name" gorm:"column:name"`
	Text      string `json:"text" gorm:"column:text"`
	Timestamp int    `json:"timestamp" gorm:"column:timestamp"`
	Tag       string `json:"tag" gorm:"column:tag"`
	Quote     int    `json:"quote" gorm:"column:quote"`
	MediaIds  string `json:"media_ids" gorm:"column:media_ids"`
}

func (Comment) TableName() string {
	return "comments"
}
