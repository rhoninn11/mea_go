package components

type ModalOpener struct {
	IDName        string
	Target        string
	LinkToContent string
}
type ModalDesc struct {
	Hid      HtmxId
	WithLink string
}

type ActionLink struct {
	IDName       string
	Target       string
	LinkToAction string
}

type HtmxId struct {
	JustName string
	TargName string
}
