package carry

type CommitReader interface {
	Read() ([]*CommitSummary, error)
}

func NewReaderFromFile(fpath, overrides string) (CommitReader, error) {
	overrider, err := newOverrider(overrides)
	if err != nil {
		return nil, err
	}

	return &carry{
		reader:    &csvReader{fpath: fpath},
		overrider: overrider,
	}, nil
}

type carry struct {
	reader    CommitReader
	overrider Overrider
}

func (c *carry) Read() ([]*CommitSummary, error) {
	commits, err := c.reader.Read()
	if err != nil {
		return nil, err
	}

	// apply override, before we start processing
	c.overrider.Override(commits)
	return commits, nil
}
