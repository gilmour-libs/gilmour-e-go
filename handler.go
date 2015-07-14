package gilmour

type Handler interface {
	Process(*GilmourRequest, *GilmourResponse)
}
