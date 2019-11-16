package Ltypedef

type (
	//CommonReqParams 通用请求参数
	CommonReqParams struct {
		Pcode string
		Version string
		Lat float64
		Lon float64
		IP string
		DEVID string
	}

	//Exampleinfo 样例信息结构体
	Exampleinfo struct {
		ID string `json:"id"`
		Title string `json:"title"`
		TitleEn string `json:"title_en"`
		Subtitle string `json:"subtitle"`
		Brief string `json:"brief"`
		Duration string `json:"duration"`
		Area string `json:"area"`
		AreaEn string `json:"areaEn"`
		Category string `json:"category"`
		CategoryEn string `json:"categoryEn"`
		Subcategory string `json:"subcategory"`
		Imgv string `json:"imgv"`
		Imgh string `json:"imgh"`
		Director string `json:"director"`
		Actor string `json:"actor"`
		Year string `json:"year"`
		Finish int `json:"finish"`
		Episode int `json:"episode"`
		Playcount int `json:"playcount"`
		Score float64 `json:"score"`
		CreateTime string `json:"create_time"`
		UpdateTime string `json:"update_time"`
		Online int `json:"online"`
		Download int `json:"download"`
	}

	//Examples 样例集合结构体
	Examples struct {
		AllExample []Exampleinfo `json:"data"`
	}

	//ExampleIDSet id集合结构体
	ExampleIDSet struct {
		IDS []string `json:"ids"`
	}

	//ExampleConfig 样例配置结构体
	ExampleConfig struct {
		ID string `json:"id"`
		Value string `json:"value"`
	}

	//ExampleConfigSet 样例配置集合结构体
	ExampleConfigSet struct {
		Set []ExampleConfig `json:"data"`
	}
)