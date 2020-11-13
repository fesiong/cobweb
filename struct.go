package cobweb

type RequestData struct {
	Body   string
	Domain string
	Scheme string
	IP     string
	Server string
}

type Website struct {
	//ID          uint64 `json:"id" gorm:"column:id;type:bigint(20);unique;unsigned;primary_key;AUTO_INCREMENT"`
	Domain      string `json:"domain" gorm:"column:domain;type:varchar(255);not null;default:'';index:idx_domain"`
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
	Url         string `json:"-" gorm:"-"`
	Links       []Link `json:"-" gorm:"-"`
}

type Link struct {
	Title  string `json:"title"`
	Url    string `json:"url"`
	Domain string `json:"domain"`
	Scheme string `json:"scheme"`
}