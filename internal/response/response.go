package response

type ResponseInfo interface {
	Status() int
	Size() int
}
