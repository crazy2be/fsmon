package fsmon

type Handler interface {
	
}

type ModifiedHandler interface {
	Handler
	Modified(name string)
}

type DeletedHandler interface {
	Handler
	Deleted(name string)
}

// TODO, not implemented
type MovedHandler interface {
	Handler
	Moved(source string, dest string)
}
