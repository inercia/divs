package rendezvous

type DhtService struct {
	id string
}

func NewDhtService(id string) (*DhtService, error) {
	d := DhtService{
		id: id,
	}
	return &d, nil
}

func (this *DhtService) Announce(external string) error {
	return nil
}

func (this *DhtService) Discover(discoveries chan string) error {
	return nil
}

func (this *DhtService) Leave() error {
	return nil
}
