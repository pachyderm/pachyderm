package lion

var (
	discardPusherInstance = &discardPusher{}
)

type discardPusher struct{}

func (d *discardPusher) Flush() error {
	return nil
}

func (d *discardPusher) Push(_ *Entry) error {
	return nil
}
