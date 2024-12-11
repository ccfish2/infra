package cell

type group []Cell

func Group(cells ...Cell) group {
	return group(cells)
}

func (g group) Info(c container) Info {
	nd := NewInfoNode("")
	for _, cell := range g {
		nd.Add(cell.Info(c))
	}
	return nd
}

func (g group) Apply(c container) error {
	for _, cell := range g {
		if err := cell.Apply(c); err != nil {
			return err
		}
	}
	return nil
}
