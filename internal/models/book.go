package models

type Link struct {
	Name string
	Url  string
}

type Book struct {
	ID          int
	Title       string
	Author      string
	Description string
	Publisher   string
	Image       string
	AmazonURL   string
	Rank        int
	Links       []Link
}

type NYTResponse struct {
	Results struct {
		Lists []struct {
			DisplayName string `json:"display_name"`
			Books       []struct {
				Title       string `json:"title"`
				Author      string `json:"author"`
				Description string `json:"description"`
				Publisher   string `json:"publisher"`
				Image       string `json:"book_image"`
				AmazonURL   string `json:"amazon_product_url"`
				Rank        int    `json:"rank"`
				BuyLinks    []struct {
					Name string `json:"name"`
					Url  string `json:"url"`
				} `json:"buy_links"`
			} `json:"books"`
		} `json:"lists"`
	} `json:"results"`
}
