package entity

type Blog struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewBlog(title, body string) *Blog {
	return &Blog{
		Title: title,
		Body:  body,
	}
}

type CreateBlogRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type CreateBlogRequestQueue struct {
	CreateBlogRequest

	CallbackCh string `json:"callback_ch"`
}

type CreateBlogResponse struct {
	ID int64 `json:"id"`
}
