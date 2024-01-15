package cobweb

type Website struct {
	ID          int    `json:"id" gorm:"column:id;not null;PRIMARY_KEY;AUTO_INCREMENT"`
	Domain      string `json:"domain" gorm:"column:domain;type:varchar(255);not null;default:'';index:idx_domain"`
	TopDomain   string `json:"top_domain" gorm:"column:top_domain;type:varchar(255);not null;default:'';index:idx_top_domain"`
	Scheme      string `json:"scheme" gorm:"column:scheme;type:varchar(10);not null;default:''"`
	Title       string `json:"title" gorm:"column:title;type:varchar(255);not null;default:''"`
	Description string `json:"description" gorm:"column:description;type:varchar(255);not null;default:''"`
	IP          string `json:"ip" gorm:"column:ip;type:varchar(20);not null;default:''"`
	WeChat      string `json:"wechat" gorm:"column:wechat;type:varchar(100);not null;default:''"`
	QQ          string `json:"qq" gorm:"column:qq;type:varchar(100);not null;default:''"`
	Cellphone   string `json:"cellphone" gorm:"column:cellphone;type:varchar(100);not null;default:''"`
	Server      string `json:"server" gorm:"column:server;type:varchar(255);not null;default:''"`
	Cms         string `json:"cms" gorm:"column:cms;type:varchar(100);not null;default:''"`
	Status      uint   `json:"status" gorm:"column:status;type:tinyint(1);not null;default:0"`
	UpdatedTime int64  `json:"updated_time" gorm:"column:updated_time;type:int(11);autoUpdateTime;index:idx_updated_time"`
	Url         string `json:"-" gorm:"-"`
	Links       []Link `json:"-" gorm:"-"`
	// 不写入这里，而是单独一个表
	Content string `json:"content" gorm:"-"`
}

type Link struct {
	Title     string `json:"title"`
	Url       string `json:"url"`
	Domain    string `json:"domain"`
	TopDomain string `json:"top_domain"`
	Scheme    string `json:"scheme"`
}
type WebsiteData struct {
	ID      int    `json:"id" gorm:"column:id;not null;PRIMARY_KEY;AUTO_INCREMENT"`
	Content string `json:"content" gorm:"column:content;type:longtext default null"`
}
